package scan

import "time"

type Result struct {
	Host    string
	Port    int
	Open    bool
	Latency time.Duration
	Service string
	Banner  string
	Error   string
}
