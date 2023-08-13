package slackbot

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/slackbot/internal/connector"
	"github.com/slack-go/slack"
	"log/slog"
	"sync"
)

type slackClient struct {
	connector    connector.SlackConnector
	eventChannel chan *slack.MessageEvent
	logger       *slog.Logger
	connected    bool
	userID       string
	lock         sync.RWMutex
}

// newSlackClient created a new slackClient
func newSlackClient(token string, logger *slog.Logger) *slackClient {
	return &slackClient{
		connector:    connector.CreateConnector(token),
		eventChannel: make(chan *slack.MessageEvent, 20),
		logger:       logger,
	}
}

// Run starts the slackClient
func (c *slackClient) Run(ctx context.Context) {
	for {
		select {
		case msg := <-c.connector.GetIncomingEvents():
			c.processEvent(msg)
		case <-ctx.Done():
			return
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
	channels, _, err := c.connector.GetConversationsForUser(&slack.GetConversationsForUserParameters{
		Types: []string{"public_channel", "private_channel", "im"},
	})
	if err == nil {
		for _, channel := range channels {
			channelIDs = append(channelIDs, channel.ID)
		}
	}
	return channelIDs, nil
}

// Send a message to slack.  If no channel is specified, the message is broadcast to all getChannels
func (c *slackClient) Send(channelID string, attachments []slack.Attachment) error {
	return c.connector.Post(channelID, attachments)
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
	//	slog.Debug("hello")
	case *slack.ConnectedEvent:
		c.setUserID(ev.Info.User.ID)
		if !c.connected {
			c.logger.Info("connected to slack" /*, "userid", ev.Info.User.ID*/)
			c.connected = true
		}
	case *slack.MessageEvent:
		c.eventChannel <- ev
	case *slack.RTMError:
		c.logger.Error("error reading on slack RTM connection", "err", ev)
	case *slack.InvalidAuthEvent:
		c.logger.Warn("error received from slack: invalid credentials")
	}
}
