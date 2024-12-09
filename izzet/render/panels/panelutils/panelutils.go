package panelutils

import "github.com/AllenDang/cimgui-go/imgui"

var (
	TableColumn0Width float32 = 180
)

func SetupRow(label string, item func(), fillWidth bool) {
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

// func InitColumns() {
// 	imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoResize, TableColumn0Width, 0)
// }
