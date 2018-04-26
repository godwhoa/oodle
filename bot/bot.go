package bot

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	triggers   []oodle.Trigger
	commandMap map[string]oodle.Command
	client     *girc.Client
	log        *logrus.Logger
	ircClient  *IRCClient
}

func NewBot(logger *logrus.Logger, ircClient *IRCClient) *Bot {
	return &Bot{
		log:        logger,
		ircClient:  ircClient,
		commandMap: make(map[string]oodle.Command),
	}
}

// Start makes a conn., stats a readloop and uses the config
func (bot *Bot) Start() error {
	bot.log.Info("Connecting...")
	bot.ircClient.OnEvent(func(event interface{}) {
		if msg, ok := event.(oodle.Message); ok {
			bot.handleCommand(msg.Nick, msg.Msg)
		}
	})
	bot.ircClient.OnEvent(bot.relayTrigger)
	return bot.ircClient.Connect()
}

// Stop stops the bot in a graceful manner
func (bot *Bot) Stop() {
	bot.ircClient.Close()
}

func (bot *Bot) relayTrigger(event interface{}) {
	for _, trigger := range bot.triggers {
		trigger.OnEvent(event)
	}
}

func (bot *Bot) handleCommand(nick string, message string) {
	args := strings.Split(strings.TrimSpace(message), " ")
	if len(args) < 1 {
		return
	}

	// TODO: make them regular commands
	// TODO: also needs to work with custom commands
	if args[0] == ".help" && len(args) == 2 {
		if command, ok := bot.commandMap[args[1]]; ok {
			info := command.Info()
			bot.ircClient.Sendf("Desciption: %s\nUsage: %s", info.Description, info.Usage)
			return
		}
		bot.ircClient.Send("Unknown command.")
		return
	}
	if args[0] == ".list" && len(args) == 1 {
		buf := ""
		for name := range bot.commandMap {
			buf += name + ", "
		}
		buf += ".list, .help"
		bot.ircClient.Send(buf)
		return
	}

	command, ok := bot.commandMap[args[0]]
	if !ok {
		return
	}

	reply, err := command.Execute(nick, args[1:])
	switch err {
	case oodle.ErrUsage:
		bot.ircClient.Sendf("Usage: " + command.Info().Usage)
	case nil:
		bot.ircClient.Send(reply)
	default:
		bot.log.Error(err)
	}

	bot.log.WithFields(logrus.Fields{
		"cmd":    args[0],
		"caller": nick,
		"reply":  reply,
		"err":    err,
	}).Debug("CommandExec")
}

func (bot *Bot) RegisterTrigger(trigger oodle.Trigger) {
	bot.triggers = append(bot.triggers, trigger)
}

func (bot *Bot) RegisterCommand(command oodle.Command) {
	cmdinfo := command.Info()
	bot.commandMap[cmdinfo.Prefix+cmdinfo.Name] = command
}
