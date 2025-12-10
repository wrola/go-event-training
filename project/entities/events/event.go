package events

type Event interface {
	IsInternal() bool
}
