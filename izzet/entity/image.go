package entity

type ImageInfo struct {
	ImageName string
	Scale     float64
}

func NewImageInfo(name string, scale float64) *ImageInfo {
	return &ImageInfo{ImageName: name, Scale: scale}
}
