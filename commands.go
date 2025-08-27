package main

import (
	"errors"
	"fmt"

	"github.com/LouisRemes-95/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	err := s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to set user %s: %w", cmd.Args[0], err)
	}

	fmt.Printf("User set to: %s\n", cmd.Args[0])
	return nil
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
