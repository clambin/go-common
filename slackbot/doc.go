// Package slackbot provides a basic slackbot implementation.
// Using this package typically involves creating a bot as follows:
//
//	bot := slackbot.New(botName, slackToken, commands)
//	go bot.Run(context.Background())
//
// Once running, the bot will listen for any commands specified and execute them. Slackbot itself implements two commands:
// "version" (which responds with botName) and "help" (which shows all implemented commands).
//
// Applications can send messages as follows:
//
//	bot.Send(channel, []slack.Attachment{{Text: "Hello world"}})
package slackbot
