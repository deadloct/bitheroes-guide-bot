package cmd

import "github.com/bwmarrin/discordgo"

type OutputFileType string

const (
	Document OutputFileType = "document"
	Image    OutputFileType = "image"
	Text     OutputFileType = "text"
)

type GenericCommand struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Output         string         `json:"output"`
	OutputFileName string         `json:"outputFileName"`
	OutputFileType OutputFileType `json:"outputFileType"`
	Subcommands    []*GenericCommand
}

func NewGenericCommand() *GenericCommand { return &GenericCommand{} }

func (c *GenericCommand) GetName() string {
	return c.Name
}

func (c *GenericCommand) GetCommand() *discordgo.ApplicationCommand {
	cmd := &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
	}

	if len(c.Subcommands) > 0 {
		cmd.Options = make([]*discordgo.ApplicationCommandOption, 0, len(c.Subcommands))
		for _, sub := range c.Subcommands {
			cmd.Options = append(cmd.Options, c.CreateSubCommand(sub))
		}
	}

	return cmd
}

func (c *GenericCommand) CreateSubCommand(sub *GenericCommand) *discordgo.ApplicationCommandOption {
	opt := &discordgo.ApplicationCommandOption{
		Name:        sub.Name,
		Description: sub.Description,
	}

	if len(sub.Subcommands) > 0 {
		opt.Options = make([]*discordgo.ApplicationCommandOption, 0, len(c.Subcommands))
		for _, cmd := range sub.Subcommands {
			opt.Options = append(opt.Options, c.CreateSubCommand(cmd))
		}
	}

	return opt
}

func (c *GenericCommand) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &c.Output,
		// Embeds:  &response.Embeds,
	})

	return err
}
