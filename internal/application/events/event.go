package events

import "time"

type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Aggregate string      `json:"aggregate"`
	Data      interface{} `json:"data"`
	Occurred  time.Time   `json:"occurred"`
}
