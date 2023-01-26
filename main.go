package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitheroes-guide-bot/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.Info("verbose logs enabled")
	log.SetLevel(log.DebugLevel)
}

func main() {
	session, err := discordgo.New("Bot " + os.Getenv("BITHEROES_GUIDE_BOT_AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// Listen for server messages only
	session.Identify.Intents = discordgo.IntentsGuildMessages
	commandManager := cmd.NewCommandManager(session)
	if err := commandManager.Start(); err != nil {
		log.Panic(err)
	}
	defer commandManager.Stop()

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("Bot exiting...")
}
