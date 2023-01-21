package cmd

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Command interface {
	GetCommand() *discordgo.ApplicationCommand
	GetName() string
	Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error
}

type CommandManager struct {
	commands map[string]Command
	session  *discordgo.Session
}

func NewCommandManager(s *discordgo.Session) *CommandManager {
	return &CommandManager{commands: make(map[string]Command), session: s}
}

func (cm *CommandManager) LoadFromJSON(data []byte) {
	var cmds []*GenericCommand
	if err := json.Unmarshal(data, &cmds); err != nil {
		log.Fatal("Error during JSON unmarshall: ", err)
	}

	if len(cmds) == 0 {
		log.Fatal("No commmands to load, exiting...")
	}

	for _, cmd := range cmds {
		cm.commands[cmd.GetName()] = cmd
	}
}

func (cm *CommandManager) Start() error {
	// SlashCommands command handler
	cm.session.AddHandler(cm.commandHandler)

	// Open up the session
	if err := cm.session.Open(); err != nil {
		return err
	}

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

func (cm *CommandManager) commandHandler(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Loading results..."},
	})

	name := i.ApplicationCommandData().Name
	cmd, ok := cm.commands[name]
	if !ok {
		log.Error("unsupported command %s received", name)
		return
	}

	if err := cmd.Handle(sess, i); err != nil {
		log.Error(err)
		errstr := "There was an error loading that command, please try again later."
		sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errstr,
		})
	}
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
