package slackbot

import (
	"context"
	"github.com/clambin/go-common/slackbot/internal/connector"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"sync"
	"testing"
	"time"
)

func TestClient_Receive(t *testing.T) {
	c := newSlackClient("dummy_token", slog.Default())
	f := connector.NewFakeConnector()
	c.connector = f

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(ctx)
	}()

	_, err := c.GetUserID()
	assert.Error(t, err)

	f.Connect()

	require.Eventually(t, func() bool {
		userId, err := c.GetUserID()
		return err == nil && userId == "123"
	}, time.Second, time.Millisecond)

	f.IncomingMessage("", "hello")

	msg := <-c.GetMessage()
	assert.Equal(t, "hello", msg.Text)

	channels, err := c.GetChannels()
	require.NoError(t, err)
	assert.Equal(t, []string{"123"}, channels)

	cancel()
	wg.Wait()
}

func TestClient_Send(t *testing.T) {
	c := newSlackClient("dummy_token", slog.Default())
	f := connector.NewFakeConnector()
	c.connector = f

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(ctx)
	}()

	f.Connect()

	require.Eventually(t, func() bool {
		userId, err := c.GetUserID()
		return err == nil && userId == "123"
	}, time.Second, time.Millisecond)

	err := c.Send("123", []slack.Attachment{{Title: "foo", Text: "bar"}})
	require.NoError(t, err)
	msg := <-f.ToSlack
	assert.Equal(t, connector.PostedMessage{
		ChannelID:   "123",
		Attachments: []slack.Attachment{{Title: "foo", Text: "bar"}},
	}, msg)

	cancel()
	wg.Wait()
}
