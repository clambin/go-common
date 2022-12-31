package client

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"golang.org/x/exp/slog"
	"sync"
)

// SlackClient interface for a slackClient
//
//go:generate mockery --name SlackClient
type SlackClient interface {
	Run(ctx context.Context)
	Send(channel string, attachments []slack.Attachment) (err error)
	GetMessage() chan *slack.MessageEvent
	GetChannels() (channelIDs []string, err error)
	GetUserID() (string, error)
}

type slackClient struct {
	slackClient  *slack.Client
	slackRTM     *slack.RTM
	eventChannel chan *slack.MessageEvent
	connected    bool
	userID       string
	lock         sync.RWMutex
}

// New created a new slackClient
func New(token string) SlackClient {
	return &slackClient{
		slackClient:  slack.New(token),
		eventChannel: make(chan *slack.MessageEvent, 20),
	}
}

// Run starts the slackClient
func (c *slackClient) Run(ctx context.Context) {
	c.slackRTM = c.slackClient.NewRTM()
	go c.slackRTM.ManageConnection()
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case msg := <-c.slackRTM.IncomingEvents:
			c.processEvent(msg)
		}
	}
}

// GetMessage returns the channel that will receive message events from Slack
func (c *slackClient) GetMessage() chan *slack.MessageEvent {
	return c.eventChannel
}

// GetChannels returns all channels the bot can post on.
// This is either the bot's direct channel or any channels the bot has been invited to
func (c *slackClient) GetChannels() ([]string, error) {
	var channelIDs []string
	channels, _, err := c.slackClient.GetConversationsForUser(&slack.GetConversationsForUserParameters{
		Types: []string{"public_channel", "private_channel", "im"},
	})
	if err == nil {
		for _, channel := range channels {
			channelIDs = append(channelIDs, channel.ID)
		}
	}
	return channelIDs, nil
}

// Send a message to slack.  if no channel is specified, the message is broadcast to all getChannels
func (c *slackClient) Send(channel string, attachments []slack.Attachment) error {
	_, _, err := c.slackRTM.PostMessage(
		channel,
		slack.MsgOptionAttachments(attachments...),
		slack.MsgOptionAsUser(true),
	)
	return err
}

func (c *slackClient) setUserID(userID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.userID = userID
}

// GetUserID returns the slack user ID of the user logged into Slack. This only returns a user ID if slackClient is logged in.
// Otherwise, an error is returned
func (c *slackClient) GetUserID() (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var err error
	if c.userID == "" {
		err = fmt.Errorf("GetUserID: not connected to Slack")
	}
	return c.userID, err
}

func (c *slackClient) processEvent(event slack.RTMEvent) {
	switch ev := event.Data.(type) {
	// case *slack.HelloEvent:
	//	log.Debug("hello")
	case *slack.ConnectedEvent:
		c.setUserID(ev.Info.User.ID)
		if !c.connected {
			slog.Info("connected to slack")
			c.connected = true
		}
	case *slack.MessageEvent:
		c.eventChannel <- ev
	case *slack.RTMError:
		slog.Error("error reading on slack RTM connection", ev)
	case *slack.InvalidAuthEvent:
		slog.Warn("error received from slack: invalid credentials")
	}
	return
}
