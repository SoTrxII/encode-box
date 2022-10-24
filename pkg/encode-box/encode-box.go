package encode_box

import (
	"context"
	"encode-box/pkg/encoder"
	console_parser "encode-box/pkg/encoder/console-parser"
	"encode-box/pkg/logger"
	object_storage "encode-box/pkg/object-storage"
	"fmt"
	"math"
	"os"
	"time"
)

var (
	log = logger.Build()
)

type EncodeBoxOptions struct {
	// Number of time to retry calls made to the object store.
	// Each call will be followed by a wait time of (2^attempt)s
	ObjStoreMaxRetry int8
}
type EncodeBox[T object_storage.BindingProxy] struct {
	// Assets Downloader
	Downloader *object_storage.ObjectStorage[T]
	// Context
	Ctx context.Context
	// Cancel function
	Cancel context.CancelFunc
	// Error channel
	EChan chan error
	// Progress channel
	PChan chan console_parser.EncodingProgress
	// Behaviour options
	opt EncodeBoxOptions
}

func NewEncodeBox[T object_storage.BindingProxy](ctx *context.Context, downloader *object_storage.ObjectStorage[T], opt *EncodeBoxOptions) *EncodeBox[T] {
	eCtx, cancel := context.WithCancel(*ctx)

	return &EncodeBox[T]{
		Downloader: downloader,
		Ctx:        eCtx,
		Cancel:     cancel,
		EChan:      make(chan error),
		PChan:      make(chan console_parser.EncodingProgress),
		opt:        *opt,
	}
}

func (eb *EncodeBox[T]) Encode(req *EncodingRequest, output string) {
	defer eb.Cancel()
	log.Infof(`Now processing encoding request %+v`, req)

	// Download required assets
	allAssets := NewAssetCollectionFrom(req)
	log.Info(`Downloading required assets...`)
	err := eb.downloadAssets(allAssets)
	if err != nil {
		log.Errorf(`Error while downloading assets : %s`, err)
		eb.EChan <- err
	}
	// Queue the assets cleaning up
	defer eb.cleanUpAssets(allAssets)
	// Choose encoding method
	// If no method found -> abort
	enc, err := eb.setupEnc(req, allAssets, output)
	if err != nil {
		log.Errorf(`Error while setup encoding : %s`, err)
		eb.EChan <- err
	}

	// Finally, start the encoding process itself
	log.Debugf("Now executing FFMPEG cmd : %s", enc.GetCommandLine())
	go enc.Start()
	for {
		select {
		case p := <-enc.PChan:
			fmt.Printf("%+v\n", p)
		case e := <-enc.EChan:
			eb.EChan <- fmt.Errorf("error while encoding : %w", e)
		case <-enc.Ctx.Done():
			return
		}
	}
}

// Concurrently download all assets required for the transcoding process
// Modify in place the array pointer
func (eb *EncodeBox[T]) downloadAssets(assets *AssetCollection) error {
	errorChannel := make(chan error)
	successChannel := make(chan bool, len(*assets))

	// Fire all downloads concurrently
	for _, asset := range *assets {
		log.Debugf(`Downloading asset "%s"`, asset.key)
		go func(asset *Asset) {
			var pathPtr *string
			var err error
			for attempts := int8(0); attempts <= eb.opt.ObjStoreMaxRetry; attempts++ {
				pathPtr, err = eb.Downloader.Download(asset.key)
				if err == nil {
					break
				}
				log.Warnf("error in attempt %d at dowloading the video from the object storage %s", attempts, err.Error())
				// We will use a linear backoff strategy, we don't have any collision whatsoever, we just want to wait until the video is available
				// The sum 2^n from 0 to 10 = 2047 ~= 30min  of total wait, this is way more than enough, as more will be over an
				// http session time. Plus, if the waiting time is really because of the b64 decoding, it's a 0(n) time complexity algorithm
				delaySecs := int64(math.Pow(2, float64(attempts)))
				time.Sleep(time.Duration(delaySecs) * time.Second)
			}
			if err != nil {
				errorChannel <- err
				return
			}
			asset.path = *pathPtr
			// Mark this goroutine as succeeded
			successChannel <- true
		}(asset)
	}

	successCounter := 0
	// And wait for all of them
Loop:
	for {
		select {
		// If any download fails, abort everything
		case e := <-errorChannel:
			return fmt.Errorf("Error while downloading required assets %s", e.Error())
		case <-successChannel:
			successCounter++
			// If every asset was downloaded, break the loop and return
			if successCounter >= len(*assets) {
				break Loop
			}
		}
	}
	return nil
}

// Remove all assets form disk
func (eb *EncodeBox[T]) cleanUpAssets(assets *AssetCollection) {
	// Fire all downloads concurrently
	for i, asset := range *assets {
		log.Infof("[Encode box] :: Trying to delete asset %s", asset.path)
		err := os.Remove(asset.path)
		if err != nil {
			log.Warnf("[Encode box] :: Could not delete asset %s", asset.path)
		}
		// Reset the asset path
		(*assets)[i].path = ""
	}

}

// Setup an Encoder instance with the downloaded assets
func (eb *EncodeBox[T]) setupEnc(req *EncodingRequest, assets *AssetCollection, output string) (*encoder.Encoder, error) {
	var enc *encoder.Encoder
	var err error

	// No audio tracks, cannot proceed
	if len(req.AudiosKeys) == 0 {
		return nil, fmt.Errorf("no suitable encoder found for %+v", req)

	}
	// The encoder can be either an Audio/Video encoder
	if req.VideoKey != "" && req.ImageKey == "" {
		enc, err = encoder.GetAudiosVideoEnc(&eb.Ctx, assets.VideosPaths()[0], assets.AudiosPaths(), output)
	} else if req.ImageKey != "" && req.VideoKey == "" {
		// Or an image/video encoder
		enc, err = encoder.GetAudiosImageEnc(&eb.Ctx, assets.ImagesPaths()[0], assets.AudiosPaths(), output)
	} else if req.ImageKey == "" && req.VideoKey == "" {
		enc, err = encoder.GetAudiosOnlyEnc(&eb.Ctx, assets.AudiosPaths(), output)
	} else {
		// If an unsupported assets set is passed, don't event try and error out
		return nil, fmt.Errorf("no suitable encoder found for %+v", req)
	}
	// If any errors happened during the creation of the encoder instance, propagate it
	if err != nil {
		return nil, fmt.Errorf("error while creating encoder :  %w", err)
	}
	return enc, nil
}

type EncodingRequest struct {
	// Record UUID
	JobId string `json:"jobId"`
	// Storage backend keys for all videos tracks
	VideoKey string `json:"videoKey"`
	// Storage backend keys for all audio tracks
	AudiosKeys []string `json:"audiosKeys"`
	// Storage backend keys for the image track
	ImageKey string `json:"imageKey"`
	// All available options for encoding
	Options EncodingOptions `json:"options"`
}

// EncodingOptions All valid encoding options
type EncodingOptions struct {
	// Clean up used video/audio/images assets if the encoding succeeded
	DeleteAssetsFromObjStore bool `json:"deleteAssetsFromObjStore"`
}
