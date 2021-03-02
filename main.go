package main

import (
	"sanjaq/logger"
	"sanjaq/server"
)

func main() {
	// Simper logger
	log := logger.InitLog()

	service := server.NewServer(log)
	service.Run()
}
