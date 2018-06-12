package model

type Emoji struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type EmojiStats struct {
	Items    []Emoji `json:"emojis"`
	Visitors int     `json:"visitors"`
}
