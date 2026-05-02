package config

import (
	"os"
	"time"
)

const (
	ServiceName = "PrintAgentService"
	ServiceDisplayName = "Print Agent Service"
	ServiceDescription = "Servicio local para impresión directa desde aplicaciones web."

	DefaultAddr = "127.0.0.1:9876"

	ShutdownTimeout = 5 * time.Second
	ReadTimeout = 15 * time.Second
	WriteTimeout = 60 * time.Second
	IdleTimeout = 120 * time.Second
	ZPLCopyDelay = 200 * time.Millisecond
)

func Addr() string {
	if v := os.Getenv("PRINT_AGENT_ADDR"); v != "" {
		return v
	}
	return DefaultAddr
}
