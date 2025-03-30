package main

import (
	"log"
	"tele/internal/app"
)

func main() {
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
