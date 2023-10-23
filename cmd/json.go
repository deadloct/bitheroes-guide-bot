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

	GuideOptionKey         = "guide"
	GuideOptionDescription = "Name of the guide to display"

	ObsoleteTitle = "⚠️ **OBSOLETE** ⚠️"
)

var FilesLocation = path.Join(".", "data")

type JSONCommand struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Guides      []*JSONCommandOption `json:"guides"`
}

type JSONCommandOption struct {
	Name        string                         `json:"name"`
	Builds      []string                       `json:"builds"`
	Familiars   []string                       `json:"fams"`
	Attachments []*JSONCommandOptionAttachment `json:"attachments"`
	Obsolete    string                         `json:"obsolete,omitempty"`
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

	guideopt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "guide",
		Description: "Name of the guide to display",
		Required:    true,
	}

	for _, guide := range c.Guides {
		name := guide.Name
		if len(guide.Builds) > 0 {
			name = fmt.Sprintf("%s (%s)", guide.Name, strings.Join(guide.Builds, ", "))
		}

		guideopt.Choices = append(guideopt.Choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: guide.Name,
		})
	}

	cmd.Options = append(cmd.Options, guideopt)
	return cmd
}

func (c *JSONCommand) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	var (
		content string
		embeds  []*discordgo.MessageEmbed
		files   []*discordgo.File
	)

	params := c.getParams(i.ApplicationCommandData().Options)
	guideParam := params[0]

	if guideParam.Name != GuideOptionKey {
		return fmt.Errorf("unsupported command %s", params[0].Name)
	}

	logger.Debugf(sess, i.Interaction, "with parameter %v:%v", guideParam.Name, guideParam.Value)

	for _, guide := range c.Guides {
		if guide.Name == guideParam.StringValue() {
			content = guide.Name
			if guide.Obsolete != "" {
				// embeds = append(
				// 	[]*discordgo.MessageEmbed{{Title: ObsoleteTitle, Description: guide.Obsolete}},
				// 	embeds...,
				// )

				content = fmt.Sprintf("%v\n%v", ObsoleteTitle, guide.Obsolete)
			}

			for aNum, attachment := range guide.Attachments {
				switch attachment.AttachmentType {
				case Markdown:
					logger.Debugf(sess, i.Interaction, "posting md %s", attachment.FileName)
					v, err := ioutil.ReadFile(path.Join(FilesLocation, "responses", attachment.FileName))
					if err != nil {
						return fmt.Errorf("unable to open file %s: %w", attachment.FileName, err)
					}

					embed := &discordgo.MessageEmbed{Description: string(v[:])}
					if aNum == 0 {
						embed.Title = guide.Name
					}

					embeds = append(embeds, embed)

				case Link:
					logger.Debugf(sess, i.Interaction, "posting link %s", attachment.Link)
					content += "\n" + attachment.Link

				case File:
					filePath := path.Join(FilesLocation, "responses", attachment.FileName)
					logger.Debugf(sess, i.Interaction, "loading file %s", filePath)
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
	return fmt.Sprintf("`/%s`: %s\n", c.Name, c.Description)
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
