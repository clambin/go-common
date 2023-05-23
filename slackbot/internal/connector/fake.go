package connector

import (
	"github.com/slack-go/slack"
)

var _ SlackConnector = &FakeConnector{}

type FakeConnector struct {
	FromSlack chan slack.RTMEvent
	ToSlack   chan PostedMessage
	channels  []slack.Channel
	userID    string
	userName  string
}

type PostedMessage struct {
	ChannelID   string
	Attachments []slack.Attachment
}

func NewFakeConnector() *FakeConnector {
	return &FakeConnector{
		FromSlack: make(chan slack.RTMEvent, 1),
		ToSlack:   make(chan PostedMessage, 1),
		channels: []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{ID: "123"},
					Name:         "#foo",
				},
				IsChannel: true,
			},
		},
		userID:   "123",
		userName: "some-user",
	}
}

func (f *FakeConnector) GetIncomingEvents() chan slack.RTMEvent {
	return f.FromSlack
}

func (f *FakeConnector) GetConversationsForUser(_ *slack.GetConversationsForUserParameters) (channels []slack.Channel, nextCursor string, err error) {
	return f.channels, "unsupported", nil
}

func (f *FakeConnector) Post(channelID string, attachments []slack.Attachment) error {
	f.ToSlack <- PostedMessage{
		ChannelID:   channelID,
		Attachments: attachments,
	}
	return nil
}

func (f *FakeConnector) Connect() {
	f.FromSlack <- slack.RTMEvent{Data: &slack.ConnectedEvent{Info: &slack.Info{User: &slack.UserDetails{ID: f.userID, Name: f.userName}}}}
}

func (f *FakeConnector) IncomingMessage(channel, text string) {
	f.FromSlack <- slack.RTMEvent{Data: &slack.MessageEvent{Msg: slack.Msg{
		Channel:  channel,
		User:     f.userID,
		Text:     text,
		Username: f.userName,
	}}}
}
