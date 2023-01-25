package cmd

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	helpName        = "help"
	helpDescription = "Lists all available commands and options"
)

type Help struct {
	commands map[string]Command
}

func NewHelp(cmds map[string]Command) *Help {
	return &Help{commands: cmds}
}

func (h *Help) GetCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        helpName,
		Description: helpDescription,
	}
}

func (h *Help) GetName() string {
	return "help"
}

func (h *Help) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	content := h.Help()
	for key, cmd := range h.commands {
		if key != h.GetName() {
			content += cmd.Help()
		}
	}

	resp := &discordgo.WebhookEdit{Content: &content}
	if _, err := sess.InteractionResponseEdit(i.Interaction, resp); err != nil {
		return fmt.Errorf("could not edit response: %w", err)
	}

	return nil
}

func (h *Help) Help() string {
	return fmt.Sprintf("`/%s`: %s\n", helpName, helpDescription)
}
