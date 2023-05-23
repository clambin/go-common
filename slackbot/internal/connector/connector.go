package connector

import "github.com/slack-go/slack"

type SlackConnector interface {
	GetIncomingEvents() chan slack.RTMEvent
	GetConversationsForUser(params *slack.GetConversationsForUserParameters) (channels []slack.Channel, nextCursor string, err error)
	Post(channelID string, attachment []slack.Attachment) error
}

type slackConnector struct {
	client *slack.Client
	rtm    *slack.RTM
}

func CreateConnector(token string, options ...slack.RTMOption) SlackConnector {
	client := slack.New(token)
	c := &slackConnector{
		client: client,
		rtm:    client.NewRTM(options...),
	}
	go c.rtm.ManageConnection()
	return c
}

func (c *slackConnector) GetIncomingEvents() chan slack.RTMEvent {
	return c.rtm.IncomingEvents
}

func (c *slackConnector) GetConversationsForUser(params *slack.GetConversationsForUserParameters) (channels []slack.Channel, nextCursor string, err error) {
	return c.client.GetConversationsForUser(params)
}

func (c *slackConnector) Post(channelID string, attachments []slack.Attachment) error {
	_, _, err := c.rtm.PostMessage(
		channelID,
		slack.MsgOptionAttachments(attachments...),
		slack.MsgOptionAsUser(true),
	)
	return err
}
