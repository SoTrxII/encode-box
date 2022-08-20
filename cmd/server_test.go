package main

import (
	"bytes"
	"context"
	mock_object_storage "encode-box/internal/mock/mock-object-storage"
	encode_box "encode-box/pkg/encode-box"
	object_storage "encode-box/pkg/object-storage"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
		RecordId:   "",
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
		RecordId:   "1",
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
	fmt.Println(err.Error())
	assert.NotNil(t, err)
}

func TestMain_MakeEncodingRequest_NoAudio(t *testing.T) {
	body := bytes.Buffer{}
	eReq := encode_box.EncodingRequest{
		RecordId:   "1",
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
		RecordId:   "1",
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
		RecordId:   "1",
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

func TestMain_Healthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	healthz(w, req)
	assert.Equal(t, []byte("OK"), w.Body.Bytes())
}

func TestMain_NewEncodeRequest_Options(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	newEncodeRequest(w, req)
	assert.Equal(t, []byte("OK"), w.Body.Bytes())
}

func TestMain_NewEncodeRequest_WrongRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	newEncodeRequest(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
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
	newEncodeRequest(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
