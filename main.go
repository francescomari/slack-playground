package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/francescomari/slack-playground/internal/event"
	"github.com/francescomari/slack-playground/internal/slack"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	accessToken, ok := os.LookupEnv("SLACK_ACCESS_TOKEN")
	if !ok {
		return fmt.Errorf("SLACK_ACCESS_TOKEN not defined")
	}

	signingSecret, ok := os.LookupEnv("SLACK_SIGNING_SECRET")
	if !ok {
		return fmt.Errorf("SLACK_SIGNING_SECRET not provided")
	}

	var addr string

	flag.StringVar(&addr, "addr", ":8080", "The address to listen to")
	flag.Parse()

	slackClient := slack.Client{
		URL:         "https://slack.com",
		AccessToken: accessToken,
		HTTPClient:  http.DefaultClient,
	}

	eventCallback := callback{
		SlackClient: &slackClient,
	}

	eventHandler := event.Handler{
		SigningSecret: signingSecret,
		Callback:      &eventCallback,
	}

	http.Handle("/event", &eventHandler)

	return http.ListenAndServe(addr, nil)
}

type callback struct {
	SlackClient *slack.Client
}

func (c *callback) OnEvent(ctx context.Context, e *event.Envelope) {
	switch e.Type {
	case "event_callback":
		c.handleEventCallback(ctx, e)
	default:
		log.Printf("unsupported envelope type: %v", e.Type)
	}
}

func (c *callback) handleEventCallback(ctx context.Context, e *event.Envelope) {
	switch e.Event.Type {
	case "app_mention":
		c.handleAppMention(ctx, e)
	default:
		log.Printf("unsupported event type: %v", e.Event.Type)
	}
}

func (c *callback) handleAppMention(ctx context.Context, e *event.Envelope) {
	switch {
	case containsWord(e.Event.Text, "ping"):
		c.handleAppMentionPing(ctx, e)
	case containsWord(e.Event.Text, "help"):
		c.handleAppMentionHelp(ctx, e)
	default:
		c.handleAppMentionUnknown(ctx, e)
	}
}

func (c *callback) handleAppMentionPing(ctx context.Context, e *event.Envelope) {
	response, err := c.SlackClient.PostMessage(ctx, &slack.PostMessageRequest{
		Channel: e.Event.Channel,
		Text:    fmt.Sprintf("<@%s> pong", e.Event.User),
	})
	if err != nil {
		log.Printf("post message: %v", err)
		return
	}
	if !response.OK {
		log.Printf("post message: request failed: %v", response.Error)
		return
	}
}

func (c *callback) handleAppMentionHelp(ctx context.Context, e *event.Envelope) {
	response, err := c.SlackClient.PostMessage(ctx, &slack.PostMessageRequest{
		Channel: e.Event.Channel,
		Text:    fmt.Sprintf("<@%s> just send me a `ping`", e.Event.User),
	})
	if err != nil {
		log.Printf("post message: %v", err)
		return
	}
	if !response.OK {
		log.Printf("post message: request failed: %v", response.Error)
		return
	}
}

func (c *callback) handleAppMentionUnknown(ctx context.Context, e *event.Envelope) {
	response, err := c.SlackClient.PostMessage(ctx, &slack.PostMessageRequest{
		Channel: e.Event.Channel,
		Text:    fmt.Sprintf("<@%s> I don't understand, ask me for `help`", e.Event.User),
	})
	if err != nil {
		log.Printf("post message: %v", err)
		return
	}
	if !response.OK {
		log.Printf("post message: request failed: %v", response.Error)
		return
	}
}

func containsWord(text, word string) bool {
	word = strings.ToLower(word)

	for _, w := range strings.Split(text, " ") {
		if strings.ToLower(w) == word {
			return true
		}
	}

	return false
}
