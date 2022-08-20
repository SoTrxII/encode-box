package encode_box

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
