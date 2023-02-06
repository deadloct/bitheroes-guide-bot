package cmd

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/deadloct/bitheroes-guide-bot/lib/logger"
)

const (
	creditsName        = "credits"
	creditsDescription = "View app contributors"
)

type Credits struct{}

func NewCredits() *Credits { return &Credits{} }

func (c *Credits) GetCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        creditsName,
		Description: creditsDescription,
	}
}

func (c *Credits) GetName() string {
	return creditsName
}

func (c *Credits) Handle(sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	logger.Debugf(i.Interaction, "handling request: /%s id:%v", helpName, i.ID)

	content := `
**Bot Credits**

_Idea for Bot_: Trogburn

_Bot Coding_: BillyIdol ([Source Code on GitHub](https://github.com/deadloct/bitheroes-guide-bot))

_Data Aggregation_: BillyIdol, ShawnBond, Trogdor, and ZombieSlayer13

_Guide Authors_: a_poor_ninja, Adhesive81, Antomanz, Ballbreaker, BillyIdol, Chocomint, ChubbyDaemon, Crow, CyberMuffin, DarkHand6, Dispel1, Ember, Fyra, Goku, Gylgymesh, Hæl (aka Hael in this bot), Huen11, iWushock, Jermoshua, JoeBu, josiah_4, kruste, McSploosh, Orcaaa, PAINisGOD93, ShawnBond, Sizz, Smolder, Special_Delivery, Techno, Tolton, VesaN, WRLD_EATR, Youreprettycute, and ZombieSlayer13

Thanks to anybody else that helped but was not mentioned because I forgot!
`
	resp := &discordgo.WebhookEdit{Content: &content}
	if _, err := sess.InteractionResponseEdit(i.Interaction, resp); err != nil {
		return fmt.Errorf("could not edit response: %w", err)
	}

	return nil
}

func (c *Credits) Help() string {
	return fmt.Sprintf("`/%s`: %s\n", creditsName, creditsDescription)
}
