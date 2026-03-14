package driver

type Status string

const (
	StatusOffline Status = "OFFLINE"
	StatusOnline  Status = "ONLINE"
	StatusBusy    Status = "BUSY"
)
