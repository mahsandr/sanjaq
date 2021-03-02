package main

import (
	"sanjaq/server"
)

func main() {
	service := server.NewServer()
	service.Run()
}
