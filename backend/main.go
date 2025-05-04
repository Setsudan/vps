package main

import (
	"log"
)

func main() {
	server, err := InitServer()
	if err != nil {
		log.Fatal("Error initializing server: ", err)
	}
	log.Printf("Server running on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed: ", err)
	}
}