package encode_box

import (
	"encode-box/pkg/encoder"
	"time"
)

// Asset An asset is a file to download from the backend storage
type Asset struct {
	// storage backend key for this asset
	key string
	// Assets downloaded path
	path string
	// Either Audio/Video or Image
	media AssetMedia
}
type AssetMedia int8

const (
	Video = iota
	Audio
	Image
	SideAudio
)

// AssetCollection an enhanced array of pointer to assets
type AssetCollection []*Asset

// NewAssetCollectionFrom Build an asset collection from an encoding requets
func NewAssetCollectionFrom(req *EncodingRequest) *AssetCollection {
	var allAssets AssetCollection
	// Add the video track if defined
	if req.VideoKey != "" {
		allAssets = append(allAssets, &Asset{
			key:   req.VideoKey,
			media: Video,
		})
	}

	// All the audio tracks
	if len(req.AudiosKeys) > 0 {
		for _, aKey := range req.AudiosKeys {
			allAssets = append(allAssets, &Asset{
				key:   aKey,
				media: Audio,
			})
		}
	}

	if req.BackgroundAudioKey != "" {
		allAssets = append(allAssets, &Asset{
			key:   req.BackgroundAudioKey,
			media: SideAudio,
		})
	}

	// And finally the image track
	if req.ImageKey != "" {
		allAssets = append(allAssets, &Asset{
			key:   req.ImageKey,
			media: Image,
		})
	}
	return &allAssets
}

// VideosPaths Get all paths of all videos assets
func (ac *AssetCollection) VideosPaths() []string {
	return ac.findPaths(func(asset *Asset) bool {
		return asset.media == Video
	})

}

// AudiosPaths Get all paths of all audio assets
func (ac *AssetCollection) AudiosPaths() []string {
	return ac.findPaths(func(asset *Asset) bool {
		return asset.media == Audio
	})
}

// ImagesPaths Get all paths of all images assets
func (ac *AssetCollection) ImagesPaths() []string {
	return ac.findPaths(func(asset *Asset) bool {
		return asset.media == Image
	})
}

func (ac *AssetCollection) SideAudiosPaths() []string {
	return ac.findPaths(func(asset *Asset) bool {
		return asset.media == SideAudio
	})
}

// Return all paths of all assets satisfying the given predicate
func (ac *AssetCollection) findPaths(predicate func(*Asset) bool) []string {
	var ret []string
	for _, a := range *ac {
		if predicate(a) {
			ret = append(ret, a.path)
		}
	}
	return ret
}

func (ac *AssetCollection) getOutputDuration() time.Duration {
	var maxDur time.Duration
	// Get maximum duration through all audios or videos assets

	// The length of the final audio track is the length of the sum of all audio tracks
	for _, a := range *ac {
		if a.path != "" && a.media == Audio {
			dur, err := encoder.GetDuration(a.path)
			if err != nil {
				log.Debugf("[Encode box] :: Could not get duration for asset %s, err : %s", a.path, err)
				continue
			}
			maxDur += dur
		}
	}

	// All other asset are in one unit, so we can directly compare
	for _, a := range *ac {
		if a.path != "" && a.media != Audio && a.media != Image {
			dur, err := encoder.GetDuration(a.path)
			if err != nil {
				log.Debugf("[Encode box] :: Could not get duration for asset %s, err : %s", a.path, err)
				continue
			}
			if dur > maxDur {
				maxDur = dur
			}
		}
	}
	return maxDur
}
