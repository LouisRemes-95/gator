package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/LouisRemes-95/gator/internal/config"
	"github.com/LouisRemes-95/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	programState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	programCommands := &commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	programCommands.register("login", handlerLogin)
	programCommands.register("register", handlerRegister)

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
		os.Exit(1)
	}
}
