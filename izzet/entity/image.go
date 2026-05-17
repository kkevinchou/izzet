package entity

type ImageComponent struct {
	ImageName string
	Scale     float64
	Billboard bool
}

func NewImageComponent(name string, scale float64, billboard bool) *ImageComponent {
	return &ImageComponent{ImageName: name, Scale: scale, Billboard: billboard}
}
