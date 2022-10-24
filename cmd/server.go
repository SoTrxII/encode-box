package main

import (
	"bytes"
	"context"
	encode_box "encode-box/pkg/encode-box"
	"encode-box/pkg/logger"
	object_storage "encode-box/pkg/object-storage"
	progress_broker "encode-box/pkg/progress-broker"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// Global logger instance
	log = logger.Build()
	// Master context
	ctx = context.Background()
	// Event broker, can be nil
	broker *progress_broker.ProgressBroker[client.Client]
	// Object store instance, use to retrieve/upload assets
	objStore *object_storage.ObjectStorage[client.Client]
)

const (
	// Env variables
	OBJECT_STORE_NAME        = "OBJECT_STORE_NAME"
	PUBSUB_NAME              = "PUBSUB_NAME"
	PUBSUB_TOPIC_PROGRESS    = "PUBSUB_TOPIC_PROGRESS"
	DAPR_MAX_REQUEST_SIZE_MB = "DAPR_MAX_REQUEST_SIZE_MB"
	// HTTP port for the server
	APP_PORT = "APP_PORT"
	// GRPC port to use to communicate with DAPR
	DAPR_GRPC_PORT = "DAPR_GRPC_PORT"

	// Default values
	// Topic to send progress event into
	DefaultPubSubTopic = "encoding-state"
	// Override default max grpc request size (4MB) for dapr client
	DefaultDaprMaxRequestSize = 2500
	// Default grpc api port for dapr
	DefaultDaprGrpcPort = 50001
	DefaultAppPort      = 8080
)

// Some kind of a root DI container
type components[T object_storage.BindingProxy] struct {
	// Encode box
	eBox *encode_box.EncodeBox[T]
	// Object backend storage
	objStore *object_storage.ObjectStorage[T]
}

// Fire a new encoding
// /!\ An HTTP return code 200 will only be returned **after** the encoding is done /!\
// This function is intended to be used with a messaging service. This way, the message will
// only be deleted from the messaging service after we made sure the processing is complete
// Although it's still possible to use it in plain HTTP, you'd have to set the HTTP_SESSION
// max time to 0
func encodeSync[T object_storage.BindingProxy](w http.ResponseWriter, req *http.Request, comp components[T]) {
	// Confirm Dapr subscription
	if req.Method == http.MethodOptions {
		_, _ = w.Write([]byte("OK"))
		return
	}
	defer req.Body.Close()

	// Check the format of the encode request...

	// Do not consume the body, instead make a copy of it
	contents, _ := io.ReadAll(req.Body)
	bodyCopy := io.NopCloser(bytes.NewReader(contents))
	defer bodyCopy.Close()
	encodeRequest, err := makeEncodingRequest(bodyCopy)
	if err != nil {
		log.Warnf(`Wrong encode request received "%s" : %s `, contents, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// And launch the encoding process...
	log.Infof(`New encoding request with id "%s" received !`, encodeRequest.JobId)
	workDir, err := os.MkdirTemp("", "encode-instance")
	if err != nil {
		http.Error(w, fmt.Sprintf("can't create temp workDir : %s", err.Error()), http.StatusInternalServerError)
	}
	outputName := fmt.Sprintf("%s.mp4", encodeRequest.JobId)
	outputPath := filepath.Join(workDir, outputName)
	err, code := encode(comp.eBox, encodeRequest, outputPath)
	if err != nil {
		log.Errorf(`error while processing encode request "%+v" : %s`, *encodeRequest, err.Error())
		http.Error(w, err.Error(), code)
		return
	}

	// Once the encoding is complete, upload the resulting video on the backend object storage...
	log.Infof(`Uploading "%s" on the backend object storage`, outputPath)
	err = comp.objStore.Upload(outputPath, outputName)
	if err != nil {
		log.Errorf(`error while upload the record in the backend object storage : %s`, err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	// And clean up temp files on the container filesystem
	// Downloaded assets are already cleaned up by the encode-box itself
	log.Infof(`Removing working directory "%s" from the local filesystem`, workDir)
	err = os.RemoveAll(workDir)
	if err != nil {
		log.Warnf(`Could not remove directiory "%s" : %s`, workDir, err.Error())
	}
	log.Infof(`Processing of request with id "%s" complete !`, encodeRequest.JobId)

	// Optionally, we can also clean up the used assets from the remote object storage

	if encodeRequest.Options.DeleteAssetsFromObjStore {
		log.Infof("Removing used assets from remote object storage")
		err = cleanUpFromObjectStore(encodeRequest, comp.objStore)
		if err != nil {
			log.Warnf(err.Error())
		}
	}
	// Finally, ACK the message
	_, _ = w.Write([]byte("OK"))
}

// Health endpoint
func healthz(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Attempt to parse a body into an encoding request
func parseBody(from io.ReadCloser) (*encode_box.EncodingRequest, error) {
	// Two types of body have to be supported : a dapr event or a raw body

	// Duplicate the body to be able to read from it twice
	contents, _ := io.ReadAll(from)
	fromCopy := io.NopCloser(bytes.NewReader(contents))
	fromCopy2 := io.NopCloser(bytes.NewReader(contents))
	defer fromCopy.Close()
	defer fromCopy2.Close()

	// First, try to parse the body as a dapr event
	var dEvt DaprEvent
	err := json.NewDecoder(fromCopy).Decode(&dEvt)
	if err != nil {
		return nil, err
	}
	// If "Type" and "Topic" are in the struct, this should be a dapr event,
	// in which case the payload is in "Data"
	if dEvt.Type != "" && dEvt.Topic != "" {
		return &dEvt.Data, nil
	}

	// Else, try to parse the request as a raw encoding request
	var eReq encode_box.EncodingRequest
	err = json.NewDecoder(fromCopy2).Decode(&eReq)
	if err != nil {
		return nil, err
	}
	return &eReq, nil
}

func cleanUpFromObjectStore[T object_storage.BindingProxy](eReq *encode_box.EncodingRequest, objStore *object_storage.ObjectStorage[T]) error {
	var failures []string
	// Video
	if eReq.VideoKey != "" {
		err := objStore.Delete(eReq.VideoKey)
		if err != nil {
			failures = append(failures, eReq.VideoKey)
		}
	}
	// Audio(s)
	for _, key := range eReq.AudiosKeys {
		err := objStore.Delete(key)
		if err != nil {
			failures = append(failures, key)
		}
	}
	// Image
	if eReq.ImageKey != "" {
		err := objStore.Delete(eReq.ImageKey)
		if err != nil {
			failures = append(failures, eReq.ImageKey)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf(`failed to delete "%s" from remote object storage`, strings.Join(failures, ", "))
	}
	return nil
}

// Format a proper encoding request from a stream
func makeEncodingRequest(from io.ReadCloser) (*encode_box.EncodingRequest, error) {
	if from == nil {
		return nil, fmt.Errorf("no body provided")
	}
	eReq, err := parseBody(from)
	if err != nil {
		return nil, err
	}
	// Sanity checks
	if eReq.JobId == "" {
		return nil, fmt.Errorf("no record id provided ")
	}
	if len(eReq.AudiosKeys) == 0 {
		return nil, fmt.Errorf("no audio track provided")
	}
	return eReq, nil
}

// Fired when an error occured while encoding a video
type encodeError struct {
	// Error message
	Message string `json:message`
}

// Fire a new encoding
func encode[T object_storage.BindingProxy](eBox *encode_box.EncodeBox[T], req *encode_box.EncodingRequest, output string) (error, int) {
	// Fire the encoding and wait for it to finish/error
	go eBox.Encode(req, output)
	for {
		select {
		case e := <-eBox.EChan:
			if broker != nil {
				broker.SendProgress(progress_broker.EncodeInfos{
					JobId: req.JobId,
					State: progress_broker.Error,
					Data:  encodeError{Message: e.Error()},
				})
			}

			eBox.Cancel()
			return e, http.StatusBadRequest
		case p := <-eBox.PChan:
			fmt.Printf("%+v", p)
			if broker != nil {
				broker.SendProgress(progress_broker.EncodeInfos{
					JobId: req.JobId,
					State: progress_broker.InProgress,
					Data:  p,
				})
			}
		case <-eBox.Ctx.Done():
			if broker != nil {
				broker.SendProgress(progress_broker.EncodeInfos{
					JobId: req.JobId,
					State: progress_broker.Done,
					Data:  nil,
				})
			}
			return nil, http.StatusOK
		}
	}
}

func makeDaprClient(maxRequestSizeMB int) (*client.Client, error) {
	var opts []grpc.CallOption

	// Getting dapr grpc port. By default, its 500001
	port := DefaultDaprGrpcPort
	// But the sidecar published a env variable with the real value
	// So we can override the value if it's defined
	if envPort, err := strconv.ParseInt(os.Getenv(DAPR_GRPC_PORT), 10, 32); err != nil && envPort != 0 {
		port = int(envPort)
	}
	opts = append(opts, grpc.MaxCallRecvMsgSize(maxRequestSizeMB*1024*1024))
	conn, err := grpc.Dial(net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)),
		grpc.WithDefaultCallOptions(opts...), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	daprClient := client.NewClientWithConnection(conn)
	return &daprClient, nil
}

// Fetch all env variables, and initializes corresponding components
func loadComponents() error {
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env file detected ! ")
	}
	// First, load the object store. This is mandatory, if it's not defined, abort
	objStoreComponent := os.Getenv(OBJECT_STORE_NAME)
	if objStoreComponent == "" {
		return fmt.Errorf(`Object store component is not defined ! Aborting !`)
	}
	maxReqSizeEnv := os.Getenv(DAPR_MAX_REQUEST_SIZE_MB)
	maxRequestSize := DefaultDaprMaxRequestSize
	if i, err := strconv.ParseInt(maxReqSizeEnv, 10, 32); err == nil && i != 0 {
		maxRequestSize = int(i)
	}
	daprClient, err := makeDaprClient(maxRequestSize)
	if err != nil {
		return fmt.Errorf("cannot init dapr client : %w", err)
	}
	objStore, err = object_storage.NewDaprObjectStorage(&ctx, daprClient, objStoreComponent)
	if err != nil {
		return fmt.Errorf("cannot init object store : %w", err)
	}

	// Next, load the event broker. This is optional, the server can function without it defined
	pubSubComponent := os.Getenv(PUBSUB_NAME)
	pubSubTopic := os.Getenv(PUBSUB_TOPIC_PROGRESS)
	if pubSubComponent != "" {
		log.Info("The pubsub component is defined ! ")
		if pubSubTopic == "" {
			pubSubTopic = DefaultPubSubTopic
		}
		broker, err = progress_broker.NewProgressBroker[client.Client](&ctx, daprClient, progress_broker.NewBrokerOptions{
			Component: pubSubComponent,
			Topic:     pubSubTopic,
		})
		if err != nil {
			return fmt.Errorf("Could not create progress broker : %w. Aborting", err)
		}
	}
	return nil
}

func main() {

	err := loadComponents()
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	http.HandleFunc("/encode", func(w http.ResponseWriter, req *http.Request) {
		encodeSync(w, req, components[client.Client]{
			eBox:     encode_box.NewEncodeBox(&ctx, objStore, &encode_box.EncodeBoxOptions{ObjStoreMaxRetry: 10}),
			objStore: objStore,
		})
	})
	http.HandleFunc("/healthz", healthz)

	port := DefaultAppPort
	fmt.Println(os.Getenv(APP_PORT))
	if i, err := strconv.ParseInt(os.Getenv(APP_PORT), 10, 32); err == nil && i != 0 {
		port = int(i)
	}
	log.Infof("Started server on PORT %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

// An event as forwarded by dapr
type DaprEvent struct {
	Type  string                     `json:"type"`
	Topic string                     `json:"topic"`
	Data  encode_box.EncodingRequest `json:"data"`
}
