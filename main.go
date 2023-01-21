package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitheroes-community-bot/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("verbose logs enabled")
	log.SetLevel(log.DebugLevel)

	session, err := discordgo.New("Bot " + os.Getenv("BITHEROES_COMMUNITY_BOT_AUTH_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// Listen for server (guild) messages only
	session.Identify.Intents = discordgo.IntentsGuildMessages

	commandsJSON, err := ioutil.ReadFile("./data/commands.json")
	if err != nil {
		log.Fatal("Error when opening commands JSON: ", err)
	}

	commandManager := cmd.NewCommandManager(session)
	commandManager.LoadFromJSON(commandsJSON)
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
