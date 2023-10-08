package encode_box

import (
	"context"
	"encode-box/internal/mock/mock-object-storage"
	object_storage "encode-box/pkg/object-storage"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

const (
	ResPath = "../../resources"
)

func Setup(t *testing.T) (*mock_object_storage.MockBindingProxy, *EncodeBox) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	proxy := mock_object_storage.NewMockBindingProxy(ctrl)
	objectStore := object_storage.NewObjectStorage(&ctx, proxy, "", false)
	eBox := NewEncodeBox(&ctx, objectStore, &EncodeBoxOptions{ObjStoreMaxRetry: 0})
	return proxy, eBox
}

func TestEncodeBox_DownloadAssets(t *testing.T) {
	proxy, eBox := Setup(t)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	aCol := getAssetsCollection(1, 1, 0)
	err := eBox.downloadAssets(aCol)
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(eBox.Tmpdir, (*aCol)[0].key), (*aCol)[0].path)
	assert.Equal(t, filepath.Join(eBox.Tmpdir, (*aCol)[1].key), (*aCol)[1].path)
}

// The default asset colelction should have no path, so duuration cannot be retrieved
func TestEncodeBox_DownloadAssets_GetDuration_NoDownload(t *testing.T) {
	aCol := getAssetsCollection(1, 1, 0)
	assert.Zero(t, aCol.getOutputDuration())
}

// It should not panic on invalid data
func TestEncodeBox_DownloadAssets_GetDuration_InvalidData(t *testing.T) {
	proxy, eBox := Setup(t)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	aCol := getAssetsCollection(1, 1, 0)
	err := eBox.downloadAssets(aCol)
	assert.Nil(t, err)
	assert.Zero(t, aCol.getOutputDuration())

}

func TestEncodeBox_DownloadAssetsErr(t *testing.T) {
	proxy, eBox := Setup(t)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	aCol := getAssetsCollection(1, 1, 0)
	err := eBox.downloadAssets(aCol)
	assert.NotNil(t, err)
}

func TestEncodeBox_DownloadAssetsErrRetry(t *testing.T) {
	proxy, eBox := Setup(t)
	eBox.opt.ObjStoreMaxRetry = 3
	aCol := getAssetsCollection(1, 1, 0)
	// Using the same naming convention as the asset collection
	vidName := "V_0"
	audName := "A_0"
	// Video will only succeed at the third attempt
	proxy.EXPECT().InvokeBinding(gomock.Any(), NewBidingMatcher(vidName, "get")).Return(nil, fmt.Errorf("test"))
	proxy.EXPECT().InvokeBinding(gomock.Any(), NewBidingMatcher(vidName, "get")).Return(nil, fmt.Errorf("test"))
	proxy.EXPECT().InvokeBinding(gomock.Any(), NewBidingMatcher(vidName, "get")).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	// Audio will succeed at the second attempt
	proxy.EXPECT().InvokeBinding(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test"))
	proxy.EXPECT().InvokeBinding(gomock.Any(), NewBidingMatcher(audName, "get")).Return(&client.BindingEvent{Data: []byte("a")}, nil)
	err := eBox.downloadAssets(aCol)
	assert.Nil(t, err)
}

func TestEncodeBox_SetupEncVideoAudio_AudioVideo(t *testing.T) {
	_, eBox := Setup(t)

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

	// O video track, multiple audio track, 0 image tracks -> AudioOnlyEnv
	req = &EncodingRequest{
		VideoKey:   "",
		AudiosKeys: []string{"d"},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(0, 3, 0)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.Nil(t, err)

	// 1 video track, multiple audio track, 1 image tracks -> Error
	req = &EncodingRequest{
		VideoKey:   "a",
		AudiosKeys: []string{"d"},
		ImageKey:   "a",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(1, 3, 1)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.NotNil(t, err)

	// 1 video track, 0 audio track, 0 image tracks -> Error
	req = &EncodingRequest{
		VideoKey:   "a",
		AudiosKeys: []string{},
		ImageKey:   "",
		Options:    EncodingOptions{},
	}
	aCol = getAssetsCollection(1, 0, 0)
	_, err = eBox.setupEnc(req, aCol, "testoutput")
	assert.NotNil(t, err)

}

func TestEncodeBox_CleanUpAssets(t *testing.T) {
	_, eBox := Setup(t)
	aCol := getAssetsCollection(1, 3, 0)
	eBox.cleanUpAssets(aCol)
}

func TestEncodeBox_Encode(t *testing.T) {
	proxy, eBox := Setup(t)
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
