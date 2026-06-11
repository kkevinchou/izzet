package entity

type CameraComponent struct {
	Target     int
	CameraMode CameraMode
	FovX       float64
}

type CameraMode string

const (
	CameraModeOverShoulder = "OVERSHOULDER"
	CameraModeWideView     = "WIDEVIEW"
)
