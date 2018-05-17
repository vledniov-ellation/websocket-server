package main

type Message struct {
	Type string `json:"type"`
	Count int `json:"count"`
}

type Messages struct {
	Items []*Message `json:"emojis"`
	Visitors int `json:"visitors"`
}
