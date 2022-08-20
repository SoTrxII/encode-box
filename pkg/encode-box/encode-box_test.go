package encode_box

import (
	"context"
	"encode-box/internal/mock/mock-object-storage"
	object_storage "encode-box/pkg/object-storage"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const (
	ResPath = "../../resources"
)

func Setup(t *testing.T) (string, *mock_object_storage.MockBindingProxy, *EncodeBox[*mock_object_storage.MockBindingProxy]) {
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)
	objectStore := object_storage.NewObjectStorage[*mock_object_storage.MockBindingProxy](&ctx, dir, proxy)
	eBox := NewEncodeBox[*mock_object_storage.MockBindingProxy](&ctx, objectStore)
	return dir, proxy, eBox
}

func TestEncodeBox_DownloadAssets(t *testing.T) {
	dir, proxy, eBox := Setup(t)
	defer Teardown(t, dir)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	aCol := getAssetsCollection(1, 1, 0)
	err := eBox.downloadAssets(aCol)
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(dir, (*aCol)[0].key), (*aCol)[0].path)
	assert.Equal(t, filepath.Join(dir, (*aCol)[1].key), (*aCol)[1].path)
}

func TestEncodeBox_DownloadAssetsErr(t *testing.T) {
	dir, proxy, eBox := Setup(t)
	defer Teardown(t, dir)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	aCol := getAssetsCollection(1, 1, 0)
	err := eBox.downloadAssets(aCol)
	assert.NotNil(t, err)
}

func TestEncodeBox_SetupEncVideoAudio_AudioVideo(t *testing.T) {
	dir, _, eBox := Setup(t)
	defer Teardown(t, dir)

	// 1 video track, 1 audio track, 0 image tracks
	req := &EncodingRequest{
		VideoKey:   "a",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol := getAssetsCollection(1, 1, 0)
	_, err := eBox.setupEnc(req, aCol, "testoutput")
	assert.Nil(t, err)

	// 1 video track, multiple audio track, 0 image tracks
	req = &EncodingRequest{
		VideoKey:   "a",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(1, 3, 0)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.Nil(t, err)

	// O video track, multiple audio track, 0 image tracks -> Error
	req = &EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(0, 3, 0)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.NotNil(t, err)

	// 1 video track, multiple audio track, 1 image tracks -> Error
	req = &EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(1, 3, 1)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.NotNil(t, err)

	// 1 video track, 0 audio track, 0 image tracks -> Error
	req = &EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(1, 0, 0)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.NotNil(t, err)

}

func TestEncodeBox_CleanUpAssets(t *testing.T) {
	dir, _, eBox := Setup(t)
	defer Teardown(t, dir)
	aCol := getAssetsCollection(1, 3, 0)
	eBox.cleanUpAssets(aCol)
}

func TestEncodeBox_Encode(t *testing.T) {
	dir, proxy, eBox := Setup(t)
	defer Teardown(t, dir)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	req := &EncodingRequest{
		VideoKey:   "d",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	go eBox.Encode(req, "output")
	for {
		select {
		case p := <-eBox.PChan:
			fmt.Printf("%+v\n", p)
		case e := <-eBox.EChan:
			fmt.Printf("%s\n", e.Error())
		case <-eBox.Ctx.Done():
			return
		}
	}
}

func Teardown(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
}

// Returns an asset collection with the specified number of videos track, audio tracks and image tracks
func getAssetsCollection(vidCount int, audCount int, imgCount int) *AssetCollection {
	var aCol AssetCollection
	for i := 0; i < vidCount; i++ {
		aCol = append(aCol, &Asset{
			key:   fmt.Sprintf("V_%d", i),
			media: Video,
		})
	}

	for i := 0; i < audCount; i++ {
		aCol = append(aCol, &Asset{
			key:   fmt.Sprintf("A_%d", i),
			media: Audio,
		})
	}

	for i := 0; i < imgCount; i++ {
		aCol = append(aCol, &Asset{
			key:   fmt.Sprintf("I_%d", i),
			media: Image,
		})
	}
	return &aCol
}
