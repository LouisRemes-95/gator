package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/LouisRemes-95/gator/internal/config"
	"github.com/LouisRemes-95/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("%s command not in register", cmd.Name)
	}

	err := f(s, cmd)
	if err != nil {
		return fmt.Errorf("failed to use registered command %s: %w", cmd.Name, err)
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

func registeredCommands() *commands {
	programCommands := &commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	programCommands.register("login", handlerLogin)
	programCommands.register("register", handlerRegister)
	programCommands.register("reset", handlerReset)
	programCommands.register("users", handlerUsers)
	programCommands.register("agg", handlerAgg)

	return programCommands
}

// Command handlers

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	_, err := s.db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("Error: User does not exist in the database.")
			os.Exit(1)
		} else {
			return fmt.Errorf("failed to get user from db: %w", err)
		}
	}

	err = s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to set user %s: %w", cmd.Args[0], err)
	}

	fmt.Printf("User set to: %s\n", cmd.Args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	myParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}
	user, err := s.db.CreateUser(context.Background(), myParams)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				fmt.Println("Error: User with that name already exists!")
				os.Exit(1)
			}
		}
		return fmt.Errorf("failed to create user in db: %w", err)
	}
	fmt.Println("User created:")
	fmt.Println("ID: ", user.ID)
	fmt.Println("Name: ", user.Name)
	fmt.Println("Create at: ", user.CreatedAt)
	fmt.Println("Updated at ", user.UpdatedAt)

	err = s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to set user %s: %w", cmd.Args[0], err)
	}
	fmt.Printf("User set to: %s\n", cmd.Args[0])
	return nil
}

func handlerReset(s *state, _ command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete all users from db: %w", err)
	}
	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users from bd: %w", err)
	}
	for _, user := range users {
		msg := "* " + user.Name
		if user.Name == s.cfg.CurrentUserName {
			msg += " (current)"
		}
		fmt.Println(msg)
	}
	return nil
}

func handlerAgg(_ *state, _ command) error {
	rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch the feed: %w", err)
	}
	fmt.Printf("%+v\n", rssFeed)
	return nil
}
