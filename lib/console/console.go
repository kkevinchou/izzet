package console

import (
	"github.com/inkyblackness/imgui-go/v4"
)

var GlobalConsole *Console = &Console{}

type ConsoleItem struct {
	Command string
	Output  string
}

type Console struct {
	ConsoleHistory []*ConsoleItem
	Input          string
	HistoryPointer int

	ScrollToBottom bool
}

func (c *Console) Send() string {
	c.ConsoleHistory = append(c.ConsoleHistory, &ConsoleItem{Command: c.Input})
	command := c.Input
	c.Input = ""
	c.HistoryPointer = 0
	return command
}

func (c *Console) AdvanceHistoryCursor(delta int) string {
	c.HistoryPointer += delta
	if c.HistoryPointer > 0 {
		c.HistoryPointer = 0
	}

	if len(c.ConsoleHistory)+c.HistoryPointer < 0 {
		c.HistoryPointer = -1 * len(c.ConsoleHistory)
	}

	index := len(c.ConsoleHistory) + c.HistoryPointer

	if index == len(c.ConsoleHistory) {
		return ""
	}

	return c.ConsoleHistory[index].Command
}

func (c *Console) InputTextCallback(data imgui.InputTextCallbackData) int32 {
	if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
		if data.EventChar() == '`' {
			return 1
		}
	} else if data.EventFlag() == imgui.InputTextFlagsCallbackHistory {
		if data.EventKey() == imgui.KeyUpArrow {
			text := c.AdvanceHistoryCursor(-1)
			data.DeleteBytes(0, len(data.Buffer()))
			data.InsertBytes(0, []byte(text))
		} else if data.EventKey() == imgui.KeyDownArrow {
			text := c.AdvanceHistoryCursor(1)
			data.DeleteBytes(0, len(data.Buffer()))
			data.InsertBytes(0, []byte(text))
		}
	}
	return 0
}
