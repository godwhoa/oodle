package main

import (
	"github.com/BurntSushi/toml"
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/plugins"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	logger.SetLevel(logrus.DebugLevel)

	config := &oodle.Config{}
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		logger.Fatal(err)
	}

	db, err := gorm.Open("sqlite3", "store.sqlite")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	oodleBot := bot.NewBot(logger)
	// TODO: better way of registering plugins
	plugins.RegisterEcho(config, oodleBot)
	plugins.RegisterSeen(config, oodleBot)
	// TODO: possible persistable interface
	plugins.RegisterTell(config, db, oodleBot)

	logger.Fatal(oodleBot.Run(config))
}
