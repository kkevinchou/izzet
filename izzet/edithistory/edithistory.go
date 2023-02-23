package edithistory

type EditHistory struct {
	editList []Edit
	cursor   int // cursor tracks where in the edit history we are
}

type Edit interface {
	Undo()
	Redo()
}

func New() *EditHistory {
	return &EditHistory{
		cursor: -1,
	}
}

func (eh *EditHistory) Append(e Edit) {
	// remove edits that appear after the cursor
	eh.editList = eh.editList[:eh.cursor+1]
	eh.editList = append(eh.editList, e)
	eh.cursor += 1
}

func (eh *EditHistory) Undo() bool {
	if eh.cursor == -1 {
		return false
	}

	edit := eh.editList[eh.cursor]
	edit.Undo()
	eh.cursor -= 1
	return true
}

func (eh *EditHistory) Redo() bool {
	if eh.cursor+1 >= len(eh.editList) {
		return false
	}

	eh.cursor += 1
	edit := eh.editList[eh.cursor]
	edit.Redo()
	return true
}

func (eh *EditHistory) Clear() {
	eh.cursor = -1
	eh.editList = nil
}
