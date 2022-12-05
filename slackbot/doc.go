// Package slackbot provides a basic slackbot implementation.
// Using this package typically involves creating a bot as follows:
//
//	bot := slackbot.New(botName, slackToken, commands)
//	go bot.Run()
//
// Once running, the bot will listen for any commands specified on the channel and execute them. Slackbot itself
// implements two commands: "version" (which responds with botName) and "help" (which shows all implemented commands).
// Additional commands can be added through the commands parameter (see New & CommandFunc):
//
//	    func doHello(args ...string) []slack.Attachment {
//		       return []slack.Attachment{{Text: "hello world " + strings.Join(args, ", ")}}
//	    }
//
// The returned attachments will be sent to the Slack channel where the command was issued.
//
// Additionally, output can be sent to the Slack channel(s) using PostChannel, e.g.:
//
//	bot.GetPostChannel() <- []slack.Attachment{{Text: "Hello world"}}
package slackbot
