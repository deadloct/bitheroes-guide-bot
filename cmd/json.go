package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type SourceType string

const (
	File     SourceType = "file"
	Markdown SourceType = "markdown"
	Link     SourceType = "link"
)

type LinkType string

const (
	LinkTypeDirect LinkType = "direct"
	LinkTypePhoto  LinkType = "photo"
	LinkTypeVideo  LinkType = "video"
)

var FilesLocation = path.Join(".", "data")

type JSONCommand struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	Subcommands []*JSONCommand                  `json:"subcommands"` // will have either subcommands or options
	Options     map[string][]*JSONCommandOption `json:"options"`
}

type JSONCommandOption struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Attachments []*JSONCommandOptionAttachment `json:"attachments"`
}

// Either a file or a link
type JSONCommandOptionAttachment struct {
	FileName       string     `json:"filename"`
	AttachmentType SourceType `json:"attachmenttype"`
	ContentType    string     `json:"contenttype"`

	Link     string   `json:"link"`
	LinkType LinkType `json:"linktype"`
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

	// The top level will never have options
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

		for key, options := range sub.Options {
			subopt := &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        key,
				Description: key,
				Required:    true,
			}

			for _, opt := range options {
				subopt.Choices = append(subopt.Choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  opt.Name,
					Value: opt.Name,
				})
			}

			cmdopt.Options = append(cmdopt.Options, subopt)
		}
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
	log.Debugf("handling request: %v id:%v", options, i.ID)

	node := c.findSubJSONCommand(options, 0)
	if node == nil {
		return fmt.Errorf("could not find handler for command %s", strings.Join(options, " "))
	}

	params := c.getParams(i.ApplicationCommandData().Options)

	for _, param := range params {
		log.Debugf("with parameter: %v:%v id:%v", param.Name, param.Value, i.ID)

		// super hack to just get this working, need to flesh this out to be more generic
		switch param.Name {
		case "authors":
			guides, ok := node.Options[param.Name]
			if !ok {
				return fmt.Errorf("there are no authors for %s", strings.Join(options, " "))
			}

			for _, guide := range guides {
				if guide.Name == param.StringValue() {
					for _, attachment := range guide.Attachments {
						switch attachment.AttachmentType {
						case Markdown:
							v, err := ioutil.ReadFile(path.Join(FilesLocation, "responses", attachment.FileName))
							if err != nil {
								return fmt.Errorf("unable to open file %s: %w", attachment.FileName, err)
							}

							embeds = append(embeds, &discordgo.MessageEmbed{
								Title:       guide.Description,
								Description: string(v[:]),
							})

						case Link:
							embed := &discordgo.MessageEmbed{Title: guide.Description}

							switch attachment.LinkType {
							case LinkTypeVideo:
								embed.Video = &discordgo.MessageEmbedVideo{
									URL:    attachment.Link,
									Width:  600,
									Height: 400,
								}

							case LinkTypePhoto:
								embed.Image = &discordgo.MessageEmbedImage{
									URL: attachment.Link,
								}

							default:
								embed.Description = fmt.Sprintf(
									"[%s](%s)",
									attachment.Link,
									attachment.Link,
								)
							}

							embeds = append(embeds, embed)

						case File:
							content = guide.Description
							filePath := path.Join(FilesLocation, "responses", attachment.FileName)
							log.Debug("loading %s", filePath)
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
		}
	}

	_, err := sess.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &embeds,
		Files:   files,
	})

	return err
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

func (c *JSONCommand) getParams(options []*discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandInteractionDataOption {
	if len(options) == 0 {
		return nil
	}

	if options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		return c.getParams(options[0].Options)
	}

	return options
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
