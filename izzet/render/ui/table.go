package ui

import "github.com/AllenDang/cimgui-go/imgui"

var (
	TableColumn0Width float32          = 180
	DefaultTableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

func Table(id string, body func()) {
	TableV(id, DefaultTableFlags, body)
}

func TableV(id string, flags imgui.TableFlags, body func()) {
	if imgui.BeginTableV(id, 2, flags, imgui.Vec2{}, 0) {
		InitColumns()
		body()
		imgui.EndTable()
	}
}

func Row(label string, item func()) {
	RowV(label, item, true)
}

func RowFit(label string, item func()) {
	RowV(label, item, false)
}

func RowV(label string, item func(), fillWidth bool) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(label)
	imgui.TableNextColumn()
	imgui.PushIDStr(label)
	if fillWidth {
		imgui.PushItemWidth(-1)
	}
	item()
	if fillWidth {
		imgui.PopItemWidth()
	}
	imgui.PopID()
}

func InitColumns() {
	imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoResize, TableColumn0Width, 0)
}

func CheckboxRow(label string, value *bool) {
	Row(label, func() { imgui.Checkbox("##value", value) })
}

func LabelRow(label string, value string) {
	Row(label, func() { imgui.LabelText("##value", value) })
}

func SliderFloatRow(label string, value *float32, min float32, max float32) {
	Row(label, func() { imgui.SliderFloat("##value", value, min, max) })
}

func SliderIntRow(label string, value *int32, min int32, max int32) {
	Row(label, func() { imgui.SliderInt("##value", value, min, max) })
}
