package drawer

import "github.com/AllenDang/cimgui-go/imgui"

func renderConfirmationModal(title string, message string, open *bool, onConfirm func(), onCancel func()) {
	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if *open {
		imgui.OpenPopupStr(title)
		*open = false
	}

	if imgui.BeginPopupModalV(title, nil, imgui.WindowFlagsAlwaysAutoResize) {
		imgui.Text(message)
		imgui.Separator()
		if imgui.Button("Delete") {
			onConfirm()
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			onCancel()
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}
}
