// Package slackbot provides a basic slackbot implementation.
// Using this package typically involves creating a bot as follows:
//
//	bot := slackbot.New("some-token", slackbot.WithCommands(...)
//	go bot.Run(context.Background())
//
// Once running, the bot connects to Slack and listens for any commands and execute them. Slackbot itself implements two commands:
// "version" (which responds with the bot's name; see WithName option) and "help" (which shows all supported commands).
//
// Applications can send messages as follows:
//
//	bot.Send(channel, []slack.Attachment{{Text: "Hello world"}})
package slackbot
