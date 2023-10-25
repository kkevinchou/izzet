package events

type Event interface{}

type PlayerJoinEvent struct {
	PlayerID int
}
