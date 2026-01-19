package x

import "time"

const (
	ShutdownTimeout          int = 30 // seconds
	ReadTimeout                  = 5 * time.Second
	ReadHeaderTimeout            = 5 * time.Second
	WriteTimeout                 = 30 * time.Second
	IdleTimeout                  = 30 * time.Second
	MutexProfileFractionRate int = 10
)
