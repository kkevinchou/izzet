package entities

type ImageInfo struct {
	ImageName string
	Scale     float32
}

func NewImageInfo(name string, scale float32) *ImageInfo {
	return &ImageInfo{ImageName: name, Scale: scale}
}
