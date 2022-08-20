package main

import (
	"context"
	encode_box "encode-box/pkg/encode-box"
	"encode-box/pkg/logger"
	object_storage "encode-box/pkg/object-storage"
	progress_broker "encode-box/pkg/progress-broker"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	log    = logger.Build()
	broker *progress_broker.ProgressBroker[client.Client]
)

const (
	// HTTP port for the server
	PORT = 8080
	// Dapr component name for the target object storage solution
	ObjStoreComponent = "object-store"
	PubSubComponent   = "object-store"
	PubSubTopic       = "encoding-state"
)

func newEncodeRequest(w http.ResponseWriter, req *http.Request) {
	// Confirm Dapr subscription
	if req.Method == http.MethodOptions {
		_, _ = w.Write([]byte("OK"))
		return
	}
	defer req.Body.Close()
	encodeRequest, err := makeEncodingRequest(req.Body)
	if err != nil {
		log.Warnf(`Wrong encode request received "%+v" : %s `, req.Body, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	eBox, err := makeEncodeBox(&ctx)
	if err != nil {
		log.Errorf(`error while processing encode request "%+v" : %s`, *encodeRequest, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dir, err := os.MkdirTemp("", "encode-instance")
	if err != nil {
		http.Error(w, fmt.Sprintf("can't create temp dir : %s", err.Error()), http.StatusInternalServerError)
	}
	output := filepath.Join(dir, "out.mp4")
	err, code := encode(eBox, encodeRequest, output)
	if err != nil {
		log.Errorf(`error while processing encode request "%+v" : %s`, *encodeRequest, err.Error())
		http.Error(w, err.Error(), code)
		return
	}
	objStore, err := makeObjStorage(&ctx)
	if err != nil {
		log.Errorf(`error while creating an object storage client: %s`, err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}
	err = objStore.Upload(output, fmt.Sprintf("%s.mp4", encodeRequest.RecordId))
	if err != nil {
		log.Errorf(`error while upload the record in the backend object storage : %s`, err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("OK"))
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Format a proper encoding request from a stream
func makeEncodingRequest(from io.ReadCloser) (*encode_box.EncodingRequest, error) {
	if from == nil {
		return nil, fmt.Errorf("no body provided")
	}
	var eReq encode_box.EncodingRequest
	err := json.NewDecoder(from).Decode(&eReq)
	if err != nil {
		return nil, err
	}
	// Sanity checks
	if eReq.RecordId == "" {
		return nil, fmt.Errorf("no record id provided ")
	}
	if len(eReq.VideoKey) == 0 && len(eReq.ImageKey) == 0 {
		return nil, fmt.Errorf("no video track provided ")
	}
	if len(eReq.AudiosKeys) == 0 {
		return nil, fmt.Errorf("no audio track provided")
	}
	return &eReq, nil
}

// Make a new object storage instance
func makeObjStorage(ctx *context.Context) (*object_storage.ObjectStorage[client.Client], error) {
	objStore, err := object_storage.NewDaprObjectStorage(ctx, ObjStoreComponent)
	if err != nil {
		return nil, err
	}
	return objStore, nil
}

// Make a new encode box instance
func makeEncodeBox(ctx *context.Context) (*encode_box.EncodeBox[client.Client], error) {
	objStore, err := object_storage.NewDaprObjectStorage(ctx, ObjStoreComponent)
	if err != nil {
		return nil, err
	}
	return encode_box.NewEncodeBox[client.Client](ctx, objStore), nil
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
					RecordId:    req.RecordId,
					EncodeState: progress_broker.Error,
					Data:        e,
				})
			}

			eBox.Cancel()
			return e, http.StatusBadRequest
		case p := <-eBox.PChan:
			fmt.Printf("%+v", p)
			if broker != nil {
				broker.SendProgress(progress_broker.EncodeInfos{
					RecordId:    req.RecordId,
					EncodeState: progress_broker.InProgress,
					Data:        p,
				})
			}
		case <-eBox.Ctx.Done():
			if broker != nil {
				broker.SendProgress(progress_broker.EncodeInfos{
					RecordId:    req.RecordId,
					EncodeState: progress_broker.Done,
					Data:        nil,
				})
			}
			return nil, http.StatusOK
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	// If pubsub component is defined in env, enable the progress broker
	pubSubComponent := os.Getenv("PUBSUB_COMPONENT")
	if pubSubComponent != "" {
		ctx := context.Background()
		daprClient, err := client.NewClient()
		if err != nil {
			log.Fatal("Error loading .env file")
			return
		}
		broker, err = progress_broker.NewProgressBroker[client.Client](&ctx, &daprClient, progress_broker.NewBrokerOptions{
			Component: "",
			Topic:     "",
		})
		if err != nil {
			log.Fatal("Error loading .env file")
			return
		}
	}

	http.HandleFunc("/encode", newEncodeRequest)
	http.HandleFunc("/healthz", healthz)
	log.Infof("Started server on PORT %d", PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
	if err != nil {
		panic(err)
	}
}
