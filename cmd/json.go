package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitheroes-guide-bot/lib/logger"
)

type SourceType string

const (
	File     SourceType = "file"
	Markdown SourceType = "markdown"
	Link     SourceType = "link"
)

var FilesLocation = path.Join(".", "data")

type JSONCommand struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Subcommands []*JSONCommand       `json:"categories"` // will have either subcommands or options
	Guides      []*JSONCommandOption `json:"guides"`
}

type JSONCommandOption struct {
	Name        string                         `json:"name"`
	Attachments []*JSONCommandOptionAttachment `json:"attachments"`
}

// Either a file or a link
type JSONCommandOptionAttachment struct {
	FileName       string     `json:"filename"`
	AttachmentType SourceType `json:"attachmenttype"`
	ContentType    string     `json:"contenttype"`
	Link           string     `json:"link"`
}

func NewJSONCommand() *JSONCommand { return &JSONCommand{} }

func (c *JSONCommand) GetName() string {
	return c.Name
}

func (c *JSONCommand) GetCommand() *discordgo.ApplicationCommand {
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

func (c *JSONCommand) CreateSubCommand(sub *JSONCommand) *discordgo.ApplicationCommandOption {
	cmdopt := &discordgo.ApplicationCommandOption{
		Name:        sub.Name,
		Description: sub.Description,
	}

	// SubCommandGroups can't have options by definition
	if len(sub.Subcommands) > 0 {
		cmdopt.Type = discordgo.ApplicationCommandOptionSubCommandGroup
		cmdopt.Options = make([]*discordgo.ApplicationCommandOption, 0, len(c.Subcommands))
		for _, cmd := range sub.Subcommands {
			cmdopt.Options = append(cmdopt.Options, c.CreateSubCommand(cmd))
		}

	} else {
		cmdopt.Type = discordgo.ApplicationCommandOptionSubCommand

		guideopt := &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "guide",
			Description: "Name of the guide to display",
			Required:    true,
		}

		for _, guide := range sub.Guides {
			guideopt.Choices = append(guideopt.Choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  guide.Name,
				Value: guide.Name,
			})
		}

		cmdopt.Options = append(cmdopt.Options, guideopt)
	}

	return cmdopt
}

func (c *JSONCommand) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	var (
		content string
		embeds  []*discordgo.MessageEmbed
		files   []*discordgo.File
	)

	options := c.buildSubCommandList(i.ApplicationCommandData().Options)
	options = append([]string{i.ApplicationCommandData().Name}, options...)
	logger.Debugf(i.Interaction, "handling request: %v", options)

	node := c.findSubJSONCommand(options, 0)
	if node == nil {
		return fmt.Errorf("could not find handler for command %s", strings.Join(options, " "))
	}

	params := c.getParams(i.ApplicationCommandData().Options)
	guideParam := params[0]

	if guideParam.Name != "guide" {
		return fmt.Errorf("unsupported command %s", params[0].Name)
	}

	logger.Debugf(i.Interaction, "with parameter %v:%v", guideParam.Name, guideParam.Value)

	for _, guide := range node.Guides {
		if guide.Name == guideParam.StringValue() {
			for _, attachment := range guide.Attachments {
				switch attachment.AttachmentType {
				case Markdown:
					logger.Debugf(i.Interaction, "posting md %s", attachment.FileName)
					v, err := ioutil.ReadFile(path.Join(FilesLocation, "responses", attachment.FileName))
					if err != nil {
						return fmt.Errorf("unable to open file %s: %w", attachment.FileName, err)
					}

					embeds = append(embeds, &discordgo.MessageEmbed{
						Title:       guide.Name,
						Description: string(v[:]),
					})

				case Link:
					logger.Debugf(i.Interaction, "posting link %s", attachment.Link)
					content = guide.Name + "\n" + attachment.Link

				case File:
					content = guide.Name
					filePath := path.Join(FilesLocation, "responses", attachment.FileName)
					logger.Debugf(i.Interaction, "loading file %s", filePath)
					fi, err := os.Open(filePath)
					if err != nil {
						return fmt.Errorf("could not open file %s: %v", filePath, err)
					}
					defer fi.Close()

					dfi := &discordgo.File{
						Name:        attachment.FileName,
						ContentType: attachment.ContentType,
						Reader:      fi,
					}
					files = append(files, dfi)

				default:
					return errors.New("unsupported file type for the response of this command")
				}
			}
		}
	}

	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &embeds,
		Files:   files,
	})

	return err
}

func (c *JSONCommand) Help() string {
	return c.help(0)
}

func (c *JSONCommand) buildSubCommandList(options []*discordgo.ApplicationCommandInteractionDataOption) []string {
	if len(options) == 0 || options[0].Value != nil {
		return nil
	}

	result := []string{options[0].Name}
	childOptions := c.buildSubCommandList(options[0].Options)
	result = append(result, childOptions...)
	return result
}

func (c *JSONCommand) findSubJSONCommand(path []string, depth int) *JSONCommand {
	if len(path) == 1 && c.Name == path[0] {
		return c
	}

	if len(path) == 1 && c.Name != path[0] {
		return nil
	}

	for _, sub := range c.Subcommands {
		if node := sub.findSubJSONCommand(path[1:], depth+1); node != nil {
			return node
		}
	}

	return nil
}

func (c *JSONCommand) getParams(options []*discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandInteractionDataOption {
	if len(options) == 0 {
		return nil
	}

	if options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		return c.getParams(options[0].Options)
	}

	return options
}

func (c *JSONCommand) help(depth int) string {
	var text string

	if depth == 0 {
		text += fmt.Sprintf("`/%s`: %s\n", c.Name, c.Description)
	} else {
		text += fmt.Sprintf("%s`%s`: %s\n", strings.Repeat("\t", depth), c.Name, c.Description)
	}

	if len(c.Subcommands) > 0 {
		for _, cmd := range c.Subcommands {
			text += cmd.help(depth + 1)
		}

		return text
	}

	if len(c.Guides) > 0 {
		text += fmt.Sprintf("%sParams: `guide`\n", strings.Repeat("\t", depth+1))
	}

	return text
}
