package slack_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"log/slog"
	"sync"
)

type Client struct {
	options   []slack.RTMOption
	client    *slack.Client
	messageCh chan *slack.MessageEvent
	clientID  string
	logger    *slog.Logger
	lock      sync.RWMutex
	rtm       *slack.RTM
}

func New(token string, logger *slog.Logger, options ...slack.RTMOption) *Client {
	return &Client{
		client:    slack.New(token),
		options:   options,
		messageCh: make(chan *slack.MessageEvent),
		logger:    logger,
	}
}

func (c *Client) Run(ctx context.Context) {
	rtm := c.client.NewRTM(c.options...)
	c.setRTM(rtm)
	go rtm.ManageConnection()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-rtm.IncomingEvents:
			c.processEvent(ev)
		}
	}
}

func (c *Client) setRTM(rtm *slack.RTM) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.rtm = rtm
}

func (c *Client) getRTM() (*slack.RTM, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.rtm == nil {
		return nil, ErrNotConnected
	}
	return c.rtm, nil
}

func (c *Client) GetMessage() chan *slack.MessageEvent {
	return c.messageCh
}

func (c *Client) Send(channelID string, attachments []slack.Attachment) error {
	var err error
	channelIDs := []string{channelID}
	if channelID == "" {
		channelIDs, err = c.GetChannels()
		if err != nil {
			return fmt.Errorf("GetChannels: %w", err)
		}
	}

	for _, id := range channelIDs {
		_, _, err = c.client.PostMessage(
			id,
			slack.MsgOptionAttachments(attachments...),
			slack.MsgOptionAsUser(true),
		)
	}
	return err
}

func (c *Client) GetChannels() ([]string, error) {
	rtm, err := c.getRTM()
	if err != nil {
		return nil, err
	}
	var channelIDs []string
	channels, _, err := rtm.GetConversationsForUser(&slack.GetConversationsForUserParameters{
		Types: []string{"public_channel", "private_channel", "im"},
	})
	if err == nil {
		for _, channel := range channels {
			channelIDs = append(channelIDs, channel.ID)
		}
	}
	return channelIDs, nil
}

var ErrNotConnected = errors.New("not connected to Slack")

func (c *Client) GetUserID() (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.clientID == "" {
		return "", ErrNotConnected
	}
	return c.clientID, nil
}

func (c *Client) setUserID(clientID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.clientID = clientID
}

func (c *Client) processEvent(event slack.RTMEvent) {
	switch ev := event.Data.(type) {
	// case *slack.HelloEvent:
	//	slog.Debug("hello")
	case *slack.ConnectedEvent:
		c.setUserID(ev.Info.User.ID)
		c.logger.Info("connected to slack", "userid", ev.Info.User.ID)
	case *slack.MessageEvent:
		c.messageCh <- ev
	case *slack.RTMError:
		c.logger.Error("error reading on slack RTM connection", "err", ev)
	case *slack.InvalidAuthEvent:
		c.logger.Warn("error received from slack: invalid credentials")

	}
}
