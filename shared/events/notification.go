package events

type NotificationEvent struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Target  string `json:"target"`
	Payload []byte `json:"payload"`
}
