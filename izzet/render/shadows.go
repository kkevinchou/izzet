package render

type ShadowPassType int

const (
	ShadowPassDirectional = iota
	ShadowPassPoint
)

type ShadowPassContext struct {
	Type ShadowPassType
}
