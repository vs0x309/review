package server

import "time"

const (
	reqTimeout      = time.Second * 15
	shutdownTimeout = time.Minute
)
