package main

import (
	"bytes"
	"context"
	mock_object_storage "encode-box/internal/mock/mock-object-storage"
	encode_box "encode-box/pkg/encode-box"
	object_storage "encode-box/pkg/object-storage"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const (
	ResDir = "../resources/test"
)

func TestMain_MakeEncodingRequest_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	_, err := makeEncodingRequest(req.Body)
	fmt.Println(err.Error())
	assert.NotNil(t, err)
}

func TestMain_MakeEncodingRequest_EmptyBody(t *testing.T) {
	body := bytes.Buffer{}
	_, err := body.WriteString("")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	fmt.Println(err.Error())
	assert.NotNil(t, err)
}
func TestMain_MakeEncodingRequest_NoId(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		JobId:      "",
		VideoKey:   "a",
		AudiosKeys: []string{"&"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	fmt.Println(err.Error())
	assert.NotNil(t, err)
}

func TestMain_MakeEncodingRequest_NoVideo(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "",
		AudiosKeys: []string{"a"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	assert.Nil(t, err)
}

func TestMain_MakeEncodingRequest_NoAudio(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "a",
		AudiosKeys: nil,
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	fmt.Println(err.Error())
	assert.NotNil(t, err)
}
func TestMain_MakeEncodingRequest_Ok_AudioVideo(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "a",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	assert.Nil(t, err)
}

func TestMain_MakeEncodingRequest_Ok_AudioImage(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "a",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err = makeEncodingRequest(req.Body)
	assert.Nil(t, err)
}

func TestMain_Encode(t *testing.T) {
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	objectStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox[*mock_object_storage.MockBindingProxy](&ctx, objectStore)
	eReq := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "a",
		Options:    encode_box.EncodingOptions{},
	}
	err, _ = encode[*mock_object_storage.MockBindingProxy](eBox, &eReq, dir)

	// Data are invalid, it will fail..
	assert.NotNil(t, err)
	// .. But with a specific error
	assert.Contains(t, err.Error(), "Invalid data")
}

func TestMain_cleanUpFromObjectStoreOk(t *testing.T) {
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	objectStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eReq := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "a",
		Options:    encode_box.EncodingOptions{},
	}
	err = cleanUpFromObjectStore[*mock_object_storage.MockBindingProxy](&eReq, objectStore)
	assert.Nil(t, err)
}

func TestMain_cleanUpFromObjectStoreError(t *testing.T) {
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	objectStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eReq := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "a",
		Options:    encode_box.EncodingOptions{},
	}
	err = cleanUpFromObjectStore[*mock_object_storage.MockBindingProxy](&eReq, objectStore)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "a, b, a")
}

func TestMain_Healthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	healthz(w, req)
	assert.Equal(t, []byte("OK"), w.Body.Bytes())
}

func TestMain_NewEncodeRequest_Options(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{})
	assert.Equal(t, []byte("OK"), w.Body.Bytes())
}

func TestMain_NewEncodeRequest_WrongRequest(t *testing.T) {
	body := bytes.Buffer{}
	_, _ = body.Write([]byte("eReqContent"))
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	w := httptest.NewRecorder()
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMain_NewEncodeRequest_DaprEvent(t *testing.T) {
	body := bytes.Buffer{}
	daprEvent := DaprEvent{
		Type:  "dapr",
		Topic: "encode",
		Data: encode_box.EncodingRequest{
			JobId:      "1",
			VideoKey:   "d",
			AudiosKeys: []string{"a", "b"},
			ImageKey:   "",
			Options:    encode_box.EncodingOptions{},
		},
	}
	eReqContent, _ := json.Marshal(daprEvent)
	_, _ = body.Write(eReqContent)
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err := parseBody(req.Body)
	assert.Nil(t, err)
}

func TestMain_NewEncodeRequest_RawRequest(t *testing.T) {
	body := bytes.Buffer{}
	rawReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "d",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, _ := json.Marshal(rawReq)
	_, _ = body.Write(eReqContent)
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	_, err := parseBody(req.Body)
	assert.Nil(t, err)
}

/*
*********************************************
*			Whole request testing			*
*********************************************
 */

// Testing a full request end2end without side effects
// will require quite a lot of mocking

// Returns a mocked new encode call with a valid encoding request as a body
func getMockedEncodingRequest(eReq encode_box.EncodingRequest) (*http.Request, *httptest.ResponseRecorder, error) {
	body := bytes.Buffer{}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		return nil, nil, err
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		return nil, nil, err
	}
	req := httptest.NewRequest(http.MethodPost, "/encode", &body)
	w := httptest.NewRecorder()
	return req, w, nil
}

// Full Ok request
func TestMain_NewEncodeRequest_AudioVideo_Ok(t *testing.T) {
	const (
		vidKey = "v.mp4"
		a1Key  = "a.m4a"
		a2Key  = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   vidKey,
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// VidKey -> Sample video
	videoContent, err := os.ReadFile(filepath.Join(ResDir, "video.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(vidKey, "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(videoContent))}, nil)

	// a1Key -> Sample audio
	audio1Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio1Content))}, nil)

	// a2Key -> Sample audio
	audio2Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio2Content))}, nil)

	// Finally, mock a Ok reponse when the server will try to upload on the remote storage
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(fmt.Sprintf("%s.mp4", eReq.JobId), "create")).
		Return(&client.BindingEvent{}, nil)

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMain_NewEncodeRequest_AudioImage_Ok(t *testing.T) {
	const (
		imgKey = "i.jpg"
		a1Key  = "a.m4a"
		a2Key  = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "",
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   imgKey,
		Options:    encode_box.EncodingOptions{},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// ImgKey -> Sample image
	imgContent, err := os.ReadFile(filepath.Join(ResDir, "test.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(imgKey, "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(imgContent))}, nil)

	// a1Key -> Sample audio
	audio1Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio1Content))}, nil)

	// a2Key -> Sample audio
	audio2Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio2Content))}, nil)

	// Finally, mock a Ok reponse when the server will try to upload on the remote storage
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(fmt.Sprintf("%s.mp4", eReq.JobId), "create")).
		Return(&client.BindingEvent{}, nil)

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMain_NewEncodeRequest_AudioOnly_Ok(t *testing.T) {
	const (
		a1Key = "a.m4a"
		a2Key = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   "",
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// a1Key -> Sample audio
	audio1Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio1Content))}, nil)

	// a2Key -> Sample audio
	audio2Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio2Content))}, nil)

	// Finally, mock a Ok reponse when the server will try to upload on the remote storage
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(fmt.Sprintf("%s.mp4", eReq.JobId), "create")).
		Return(&client.BindingEvent{}, nil)

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMain_NewEncodeRequest_Ok_WithCleanup(t *testing.T) {
	const (
		vidKey = "v.mp4"
		a1Key  = "a.m4a"
		a2Key  = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   vidKey,
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{DeleteAssetsFromObjStore: true},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// VidKey -> Sample video
	videoContent, err := os.ReadFile(filepath.Join(ResDir, "video.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(vidKey, "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(videoContent))}, nil)

	// a1Key -> Sample audio
	audio1Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio1Content))}, nil)

	// a2Key -> Sample audio
	audio2Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio2Content))}, nil)

	// Finally, mock a Ok reponse when the server will try to upload on the remote storage
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(fmt.Sprintf("%s.mp4", eReq.JobId), "create")).
		Return(&client.BindingEvent{}, nil)

	// And except any number of deletion request
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("*", "delete")).
		Return(&client.BindingEvent{}, nil).
		AnyTimes()

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMain_NewEncodeRequest_EncodingError(t *testing.T) {
	const (
		vidKey = "v.mp4"
		a1Key  = "a.m4a"
		a2Key  = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   vidKey,
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// Set every assets to random bytes, which will make FFMPEG error
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(vidKey, "get")).
		Return(&client.BindingEvent{Data: []byte("a")}, nil)

	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte("a")}, nil)

	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte("a")}, nil)

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Error during upload
func TestMain_NewEncodeRequest_UploadError(t *testing.T) {
	const (
		vidKey = "v.mp4"
		a1Key  = "a.m4a"
		a2Key  = "b.m4a"
	)
	eReq := encode_box.EncodingRequest{
		JobId:      "1",
		VideoKey:   vidKey,
		AudiosKeys: []string{a1Key, a2Key},
		ImageKey:   "",
		Options:    encode_box.EncodingOptions{},
	}
	req, w, err := getMockedEncodingRequest(eReq)
	if err != nil {
		t.Fatal(err)
	}
	// Redirect calls to the backend storage to valid assets for each required assets in the request
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)

	// VidKey -> Sample video
	videoContent, err := os.ReadFile(filepath.Join(ResDir, "video.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(vidKey, "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(videoContent))}, nil)

	// a1Key -> Sample audio
	audio1Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.
		EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("a.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio1Content))}, nil)

	// a2Key -> Sample audio
	audio2Content, err := os.ReadFile(filepath.Join(ResDir, "audio.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher("b.m4a", "get")).
		Return(&client.BindingEvent{Data: []byte(base64.StdEncoding.EncodeToString(audio2Content))}, nil)

	// Finally, mock a Ok reponse when the server will try to upload on the remote storage
	proxy.EXPECT().
		InvokeBinding(gomock.Any(), NewBidingMatcher(fmt.Sprintf("%s.mp4", eReq.JobId), "create")).
		Return(&client.BindingEvent{}, fmt.Errorf("test"))
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	objStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := encode_box.NewEncodeBox(&ctx, objStore)
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{
		eBox:     eBox,
		objStore: objStore,
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMain_NewEncodeRequest_NoDapr(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"a", "b"},
		ImageKey:   "a",
		Options:    encode_box.EncodingOptions{},
	}
	eReqContent, err := json.Marshal(eReq)
	if err != nil {
		t.Fatal(err)
	}
	_, err = body.Write(eReqContent)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	w := httptest.NewRecorder()
	encodeSync(w, req, components[*mock_object_storage.MockBindingProxy]{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

type bindingMatcher struct {
	name      string
	operation string
}

func NewBidingMatcher(name, operation string) gomock.Matcher {
	return &bindingMatcher{name, operation}
}

func (m *bindingMatcher) Matches(x interface{}) bool {
	req, ok := x.(*client.InvokeBindingRequest)
	if !ok {
		return false
	}
	if req.Operation != m.operation {
		return false
	}
	// If the name is a wildcard, accept anything
	if m.name != "*" && req.Metadata["key"] != m.name {
		return false
	}
	return true
}
func (m *bindingMatcher) String() string {
	return m.name
}
