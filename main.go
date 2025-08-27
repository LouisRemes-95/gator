package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LouisRemes-95/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	programState := &state{
		cfg: &cfg,
	}

	programCommands := &commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	programCommands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("Not enough input arguments")
		os.Exit(1)
	}

	requestedCommand := command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = programCommands.run(programState, requestedCommand)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
