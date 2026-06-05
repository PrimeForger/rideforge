package realtime

type IncomingMessage struct {
	Type string `json:"type"`

	Lat      float64 `json:"lat,omitempty"`
	Lng      float64 `json:"lng,omitempty"`
	Accuracy float64 `json:"accuracy,omitempty"`
	Speed    float64 `json:"speed,omitempty"`
	Bearing  float64 `json:"bearing,omitempty"`
	Seq      int64   `json:"seq,omitempty"`
}

type OutgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
