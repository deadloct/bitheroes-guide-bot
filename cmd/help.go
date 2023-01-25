package cmd

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

const (
	helpName        = "help"
	helpDescription = "Lists all available commands and options"
)

type Help struct {
	commands    map[string]Command
	outputOrder []string
}

func NewHelp(cmds map[string]Command) *Help {
	oo := []string{helpName}
	for key := range cmds {
		oo = append(oo, key)
	}

	sort.Strings(oo)

	return &Help{commands: cmds, outputOrder: oo}
}

func (h *Help) GetCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        helpName,
		Description: helpDescription,
	}
}

func (h *Help) GetName() string {
	return helpName
}

func (h *Help) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	var content string

	for _, key := range h.outputOrder {
		content += h.commands[key].Help()
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
