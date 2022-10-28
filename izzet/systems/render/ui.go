package render

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/playercommand/protogen/playercommand"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/izzet/lib/console"
)

func (s *RenderSystem) debugWindow() {
	imgui.SetNextWindowBgAlpha(0.5)
	imgui.BeginV("Debug", nil, imgui.WindowFlagsNoFocusOnAppearing|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove)
	s.generalInfoComponent()
	s.networkInfoUIComponent()
	s.entityInfoUIComponent()
	s.serverStatsInfoComponent()
	if imgui.IsWindowFocused() {
		s.world.SetFocusedWindow(types.WindowDebug)
	}
	imgui.End()
}

func (s *RenderSystem) consoleWindow() {
	imgui.BeginV("Console", nil, imgui.WindowFlagsNoTitleBar)

	imgui.PushItemWidth(-1)
	imgui.PushStyleColor(imgui.StyleColorFrameBg, imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1})
	for _, consoleItem := range console.GlobalConsole.ConsoleHistory {
		imgui.Textf("%s", consoleItem.Command)
	}
	imgui.PopStyleColor()
	imgui.Separator()

	flags := imgui.InputTextFlagsEnterReturnsTrue | imgui.InputTextFlagsCallbackCharFilter | imgui.InputTextFlagsCallbackHistory
	value := imgui.InputTextV("input", &console.GlobalConsole.Input, flags, console.GlobalConsole.InputTextCallback)

	if console.GlobalConsole.ScrollToBottom {
		imgui.SetScrollHereY(1)
		console.GlobalConsole.ScrollToBottom = false
	}

	if value {
		command := console.GlobalConsole.Send()
		console.GlobalConsole.ScrollToBottom = true
		s.world.GetEventBroker().Broadcast(&events.RPCEvent{Command: command})
		imgui.SetKeyboardFocusHereV(-1)
	}
	imgui.PopItemWidth()

	if imgui.IsWindowFocused() {
		s.world.SetFocusedWindow(types.WindowConsole)
	}

	for _, e := range s.events {
		if e.Type() == events.EventTypeConsoleEnabled {
			imgui.SetKeyboardFocusHereV(-1)
			break
		}
	}

	imgui.End()
}

func (s *RenderSystem) inventoryWindow() {
	player := s.world.GetPlayerEntity()
	cc := player.GetComponentContainer()
	inventoryComponent := cc.InventoryComponent
	assetManager := directory.GetDirectory().AssetManager()
	f := assetManager.GetTexture("front")
	_ = f

	imgui.SetNextWindowBgAlpha(1)
	imgui.BeginV("Inventory", nil, imgui.WindowFlagsNoFocusOnAppearing|imgui.WindowFlagsNoFocusOnAppearing|imgui.WindowFlagsNoCollapse)
	for i, slot := range inventoryComponent.Data.Items {
		imgui.PushID(fmt.Sprintf("%d", i))
		if i%inventoryComponent.Width != 0 {
			imgui.SameLine()
		}

		if slot.Id != -1 {
			label := fmt.Sprintf("money\n%dx", slot.Count)
			imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{X: 0, Y: 0.4, Z: 0.2, W: 1})
			imgui.ButtonV(label, imgui.Vec2{X: 64, Y: 64})
			imgui.PopStyleColor()
		} else {
			imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{X: 1, Y: 0.5, Z: 0, W: 1})
			imgui.ButtonV("empty", imgui.Vec2{X: 64, Y: 64})
			imgui.PopStyleColor()
			// ui.ImageButtonV(imgui.TextureID(uintptr(f.ID)), imgui.Vec2{X: 64, Y: 64}, imgui.Vec2{}, imgui.Vec2{X: 1, Y: -1}, 0, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
		}

		if imgui.BeginDragDropSource(imgui.DragDropFlagsNone) {
			str := fmt.Sprintf("%d", i)
			imgui.SetDragDropPayload("slot", []byte(str), imgui.ConditionNone)
			imgui.EndDragDropSource()
		}
		if imgui.BeginDragDropTarget() {
			if payload := imgui.AcceptDragDropPayload("slot", imgui.DragDropFlagsNone); payload != nil {
				index, err := strconv.Atoi(string(payload))
				if err != nil {
					panic(err)
				}

				itemSwap := playercommand.Wrapper_Itemswap{Itemswap: &playercommand.ItemSwap{Idx1: int64(index), Idx2: int64(i)}}
				command := playercommand.Wrapper{Playercommand: &itemSwap}
				s.world.GetEventBroker().Broadcast(&events.PlayerCommandEvent{
					Command: &command,
				})
			}
		}
		imgui.PopID()
	}

	if imgui.IsWindowFocused() {
		s.world.SetFocusedWindow(types.WindowInventory)
	}
	imgui.End()
}

func (s *RenderSystem) networkInfoUIComponent() {
	metricsRegistry := s.world.MetricsRegistry()
	predictionMiss := int(metricsRegistry.GetOneSecondSum("predictionMiss"))
	// serverPosition := int(metricsRegistry.GetLatest("serverPositionDiff"))
	predictionHit := int(metricsRegistry.GetOneSecondSum("predictionHit"))
	ping := int(metricsRegistry.GetOneSecondAverage("ping"))
	// updateMessageSize := int(metricsRegistry.GetOneSecondSum("update_message_size")) / 1000
	// updateCount := int(metricsRegistry.GetOneSecondSum("update_message_count"))
	// newInput := int(metricsRegistry.GetOneSecondSum("newinput"))

	if imgui.CollapsingHeaderV("Network", imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)
		uiTableRow("Ping", ping)
		uiTableRow("Predictions Hit", predictionHit)
		uiTableRow("Predictions Miss", predictionMiss)
		// uiTableRow("Server Position", serverPosition)
		// uiTableRow("Update Count", updateCount)
		// uiTableRow("Update Size", updateMessageSize)
		// uiTableRow("Inputs Sent", newInput)
		imgui.EndTable()
	}
}

func (s *RenderSystem) lightingUIComponent(textureID uint32) {
	if imgui.CollapsingHeaderV("Lighting", imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen) {
		imgui.ImageV(imgui.TextureID(textureID), imgui.Vec2{X: 160, Y: 90}, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
	}
}

func (s *RenderSystem) entityInfoUIComponent() {
	entity := s.world.GetPlayerEntity()
	cc := entity.GetComponentContainer()
	entityPosition := cc.TransformComponent.Position
	orientation := cc.TransformComponent.Orientation
	velocity := cc.MovementComponent.Velocity

	if imgui.CollapsingHeaderV("Entity", imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)
		uiTableRow("ID", entity.GetID())
		uiTableRow("Position", utils.PPrintVec(entityPosition))
		uiTableRow("Velocity", utils.PPrintVec(velocity))
		uiTableRow("Orientation", utils.PPrintQuatAsVec(orientation))
		imgui.EndTable()
	}
}

func (s *RenderSystem) generalInfoComponent() {
	metricsRegistry := s.world.MetricsRegistry()
	fps := int(metricsRegistry.GetOneSecondSum("fps"))
	// frameCatchup := int(metricsRegistry.GetOneSecondSum("frameCatchup"))
	if imgui.CollapsingHeaderV("General", imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)
		uiTableRow("FPS", fps)
		uiTableRow("frametime", fmt.Sprintf("%.3f", s.world.MetricsRegistry().GetOneSecondAverage("frametime")))
		uiTableRow("rendertime", fmt.Sprintf("%.3f", s.world.MetricsRegistry().GetOneSecondAverage("rendertime")))
		// uiTableRow("Frame Catchup", frameCatchup)
		uiTableRow("CF", s.world.CommandFrame())
		imgui.EndTable()
	}
}

func (s *RenderSystem) serverStatsInfoComponent() {
	serverStats := s.world.ServerStats()
	if imgui.CollapsingHeaderV("Server Stats", imgui.TreeNodeFlagsCollapsingHeader|imgui.TreeNodeFlagsDefaultOpen) {
		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)

		keys := make([]string, 0, len(serverStats))
		for k := range serverStats {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			uiTableRow(k, serverStats[k])
		}
		imgui.EndTable()
	}
}

func uiTableRow(label string, value any) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(label)
	imgui.TableSetColumnIndex(1)
	imgui.Text(fmt.Sprintf("%v", value))
}
