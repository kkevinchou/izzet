package entity

type CameraComponent struct {
	Target     int
	CameraMode CameraMode
}

type CameraMode string

const (
	CameraModeOverShoulder = "OVERSHOULDER"
	CameraModeWideView     = "WIDEVIEW"
)
