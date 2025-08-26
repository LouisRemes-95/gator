package main

import (
	"fmt"
	"log"

	"github.com/LouisRemes-95/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	err = cfg.SetUser("Louis")
	if err != nil {
		log.Fatal("Failed to set user:", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	fmt.Print(cfg)
}
