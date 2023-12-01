package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitheroes-guide-bot/lib/logger"
	log "github.com/sirupsen/logrus"
)

type Command interface {
	GetCommand() *discordgo.ApplicationCommand
	GetName() string
	Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error
	Help() string
}

type CommandManager struct {
	commands map[string]Command
	session  *discordgo.Session
}

func NewCommandManager(s *discordgo.Session) *CommandManager {
	cm := &CommandManager{commands: make(map[string]Command), session: s}
	cm.loadFromJSON()

	credits := NewCredits()
	cm.commands[credits.GetName()] = credits

	// Add help last so it gets all the other commands
	h := NewHelp(cm.commands)
	cm.commands[h.GetName()] = h

	return cm
}

func (cm *CommandManager) Start() error {
	// SlashCommands command handler
	cm.session.AddHandler(cm.commandHandler)

	// Open up the session
	if err := cm.session.Open(); err != nil {
		return err
	}

	// Just in case the app crashed last time without removing commands
	cm.cleanupCommands()

	// Add new commands
	log.Debug("registering slash commands")
	for _, c := range cm.commands {
		cmd := c.GetCommand()
		log.Debugf("creating command %v", c.GetName())
		if _, err := cm.session.ApplicationCommandCreate(cm.session.State.User.ID, "", cmd); err != nil {
			log.Panicf("cannot create command %v: %v", c.GetName(), err)
		}

		log.Debugf("created command %v", c.GetName())
	}

	log.Debug("finished registering slash commands")
	return nil
}

func (cm *CommandManager) Stop() {
	cm.cleanupCommands()
}

func (cm *CommandManager) cleanupCommands() {
	existingCommands, err := cm.session.ApplicationCommands(cm.session.State.User.ID, "")
	if err != nil {
		log.Errorf("could not retrieve commands to do a pre-startup cleanup")
	}

	log.Debug("cleaning up old slash commands...")
	for _, v := range existingCommands {
		log.Debugf("removing command %v", v.Name)
		if err := cm.session.ApplicationCommandDelete(cm.session.State.User.ID, "", v.ID); err != nil {
			log.Debugf("unable to remove command %v: %v", v.Name, err)
		} else {
			log.Debugf("removed command %v", v.Name)
		}
	}

	log.Debug("finished old command cleanup")
}

func (cm *CommandManager) commandHandler(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Loading results..."},
	})

	name := i.ApplicationCommandData().Name
	logger.Infof(sess, i.Interaction, "handling command %v", name)

	cmd, ok := cm.commands[name]
	if !ok {
		logger.Errorf(sess, i.Interaction, "unsupported command %s", name)
		return
	}

	if err := cmd.Handle(sess, i); err != nil {
		logger.Error(sess, i.Interaction, err)
		errstr := "There was an error loading that command, please try again later."
		sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errstr,
		})
	}
}

func (cm *CommandManager) loadFromJSON() {
	data, err := ioutil.ReadFile("./data/commands.json")
	if err != nil {
		log.Fatal("error when opening commands JSON: ", err)
	}

	var cmds []*JSONCommand
	if err := json.Unmarshal(data, &cmds); err != nil {
		log.Fatal("error during JSON unmarshall: ", err)
	}

	if len(cmds) == 0 {
		log.Fatal("no commmands to load, exiting...")
	}

	for _, cmd := range cmds {
		cm.commands[cmd.GetName()] = cmd
	}
}
