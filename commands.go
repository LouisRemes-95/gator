package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
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
	programCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	programCommands.register("feeds", handlerFeeds)
	programCommands.register("follow", middlewareLoggedIn(handlerFollow))
	programCommands.register("following", middlewareLoggedIn(handlerFollowing))
	programCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	programCommands.register("browse", middlewareLoggedIn(handlerBrowse))

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

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	timeBetweenReps, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	fmt.Println("Collecting feeds every %w", timeBetweenReps)
	ticker := time.NewTicker(timeBetweenReps)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Println("failed to scrape feeds", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("command arg's slice has less than 2 elements")
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get the current user: %w", err)
	}

	myParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), myParams)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				fmt.Println("Error: Feed with that url already exists!")
				os.Exit(1)
			}
		}
		return fmt.Errorf("failed to create user in db: %w", err)
	}
	fmt.Println("Feed created:")
	fmt.Println("ID: ", feed.ID)
	fmt.Println("Name: ", feed.Name)
	fmt.Println("Create at: ", feed.CreatedAt)
	fmt.Println("Updated at ", feed.UpdatedAt)
	fmt.Println("Url: ", feed.Url)
	fmt.Println("UserID: ", feed.UserID)

	FeedFollowParams := database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), FeedFollowParams)
	if err != nil {
		return fmt.Errorf("failed to create a feedfollow entry: %w", err)
	}

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Println("Name: ", feed.FeedName)
		fmt.Println("Url: ", feed.FeedUrl)
		fmt.Println("User: ", feed.UserName)
		fmt.Println("")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %w", err)
	}

	myParams := database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	feedFolow, err := s.db.CreateFeedFollow(context.Background(), myParams)
	if err != nil {
		return fmt.Errorf("failed to create a feedfollow entry: %w", err)
	}

	fmt.Println("Feed follow created:")
	fmt.Println("User name: ", feedFolow.UserName)
	fmt.Println("Feed name: ", feedFolow.FeedName)
	return nil
}

func handlerFollowing(s *state, _ command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get follows for current user: %w", err)
	}

	fmt.Println(user.Name, "follows:")
	for _, follow := range follows {
		fmt.Println(follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) == 0 {
		return errors.New("command arg's slice empty")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to get feed bu url: %w", err)
	}

	myParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), myParams)
	if err != nil {
		return fmt.Errorf("failed to delete the feedfollow entry")
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int
	err := errors.New("no input arguments")
	if len(cmd.Args) > 0 {
		limit, err = strconv.Atoi(cmd.Args[0])
	}
	if err != nil {
		println("No proper limit number of posts provided, defaulting to 2")
		print("\n")
		limit = 2
	}

	myParams := database.GetPostsforUserParams{
		Name:  user.Name,
		Limit: int32(limit),
	}

	posts, err := s.db.GetPostsforUser(context.Background(), myParams)
	if err != nil {
		return fmt.Errorf("failed to get posts for user %s: %w", user.Name, err)
	}

	for _, post := range posts {
		fmt.Println("Post title: ", post.Title)
		fmt.Println("Created at: ", post.CreatedAt)
		fmt.Println("Updated at: ", post.UpdatedAt)
		fmt.Println("Published at: ", post.PublishedAt)
		fmt.Println("Url: ", post.Url)
		fmt.Println("Description: ", post.Description)
		print("\n")
	}
	return nil
}

// Helper functions
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to get the current user: %w", err)
		}
		err = handler(s, cmd, user)
		if err != nil {
			return fmt.Errorf("handler failed: %w", err)
		}
		return err
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the next feed to fetch: %w", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("failed to mark the fetched feed: %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch the feed: %w", err)
	}

	for _, item := range rssFeed.Channel.Item {
		publishTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return fmt.Errorf("failed to parse PubDate: %w", err)
		}
		description := sql.NullString{
			String: item.Description,
			Valid:  item.Description != "", // or some other check for "is it meaningful?"
		}

		myParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: publishTime,
			FeedID:      feed.ID,
		}
		err = s.db.CreatePost(context.Background(), myParams)
		if err != nil {
			return fmt.Errorf("failed to create a post: %w", err)
		}
	}
	return nil
}
