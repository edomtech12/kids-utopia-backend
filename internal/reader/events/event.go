package events

import "time"

type EventType string

const (
	ProgressUpdated EventType = "progress.updated"
	SessionStarted  EventType = "session.started"
	SessionEnded    EventType = "session.ended"
)

type Event struct {
	EventID   string
	Type      EventType
	SessionID string
	UserID    string
	ChildID   string
	BookID    string
	Page      int
	Timestamp time.Time
}