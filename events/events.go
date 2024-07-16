package events

type EventType int

const (
	OpenCommand EventType = iota
	CloseCommand
	OpenTTS
	CloseTTS
	OpenReminderTTS
	WriteDB
)

type IEvent interface {
	GetType() EventType
}

type OpenCommandEvent struct {
}

func (o OpenCommandEvent) GetType() EventType {
	return OpenCommand
}

type CloseCommandEvent struct {
}

func (c CloseCommandEvent) GetType() EventType {
	return CloseCommand
}

type OpenTTSEvent struct {
}

func (o OpenTTSEvent) GetType() EventType {
	return OpenTTS
}

type CloseTTSEvent struct {
}

func (c CloseTTSEvent) GetType() EventType {
	return CloseTTS
}

type OpenReminderTTSEvent struct {
}

func (o OpenReminderTTSEvent) GetType() EventType {
	return OpenReminderTTS
}

type WriteDBEvent struct {
}

func (w WriteDBEvent) GetType() EventType {
	return WriteDB
}
