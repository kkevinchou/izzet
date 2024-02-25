package platforms

type Window interface {
	Minimized() bool
	WindowFocused() bool
	GetSize() (int, int)
	Swap()
}
