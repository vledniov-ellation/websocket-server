package model

// Emoji defines the messages struct received through the websocket
type Emoji struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// EmojiStats defines the aggregated data returned by the service via websocket
type EmojiStats struct {
	Items    []Emoji `json:"emojis"`
	Visitors int     `json:"visitors"`
}
