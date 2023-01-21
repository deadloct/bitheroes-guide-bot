package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type SourceType string

const (
	Document SourceType = "document"
	Image    SourceType = "image"
	Markdown SourceType = "markdown"
)

var FilesLocation = path.Join(".", "data")

type GenericCommand struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Title       string     `json:"title"`
	Output      string     `json:"output"`
	Source      string     `json:"source"`
	SourceType  SourceType `json:"sourceType"`
	Subcommands []*GenericCommand
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
		opt.Type = discordgo.ApplicationCommandOptionSubCommandGroup
		opt.Options = make([]*discordgo.ApplicationCommandOption, 0, len(c.Subcommands))
		for _, cmd := range sub.Subcommands {
			opt.Options = append(opt.Options, c.CreateSubCommand(cmd))
		}
	} else {
		opt.Type = discordgo.ApplicationCommandOptionSubCommand
	}

	return opt
}

func (c *GenericCommand) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	var (
		content string
		embeds  []*discordgo.MessageEmbed
	)

	options := c.buildOptionsList(i.ApplicationCommandData().Options)
	options = append([]string{i.ApplicationCommandData().Name}, options...)
	node := c.findSubGenericCommand(options, 0)
	if node == nil {
		return fmt.Errorf("could not find handler for command %s", strings.Join(options, " "))
	}

	if node.Source == "" {
		return fmt.Errorf("no output for the command %s, check json file", strings.Join(options, " "))
	}

	switch node.SourceType {
	case Markdown:
		v, err := ioutil.ReadFile(path.Join(FilesLocation, "responses", node.Source))
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", node.Source, err)
		}

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:       node.Title,
			Description: string(v[:]),
		})

	case Image:
		content = node.Description
		embeds = append(embeds, &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: node.Source,
			},
		})

	default:
		return errors.New("unsupported file type for the response of this command")
	}

	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &embeds,
	})

	return err
}

func (c *GenericCommand) buildOptionsList(options []*discordgo.ApplicationCommandInteractionDataOption) []string {
	if len(options) == 0 {
		return nil
	}

	result := []string{options[0].Name}
	childOptions := c.buildOptionsList(options[0].Options)
	result = append(result, childOptions...)
	return result
}

func (c *GenericCommand) findSubGenericCommand(path []string, depth int) *GenericCommand {
	log.Debugf("searching for %s on node:%s, depth:%d", path, c.Name, depth)
	if len(path) == 1 && c.Name == path[0] {
		return c
	}

	if len(path) == 1 && c.Name != path[0] {
		return nil
	}

	for _, sub := range c.Subcommands {
		if node := sub.findSubGenericCommand(path[1:], depth+1); node != nil {
			return node
		}
	}

	return nil
}
