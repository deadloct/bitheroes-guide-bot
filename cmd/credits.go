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
	logger.Debugf(sess, i.Interaction, "handling request: /%s id:%v", helpName, i.ID)

	content := `
**Bot Credits**

_Guide Authors_: 3riko, 5Rupees, a_poor_ninja, Adhesive81, Antomanz, Ballbreaker, BillyIdol, Bisamratte, Blanquiito, ChubbyDaemon, Chuck, Colb, Commander, Crow, CyberMuffin, DarkHand6, Dispel1, Dracaris, Dude_WTF, Eliealsamaan85, Ember, fohpo, Fyra, Gagf, Gavx, Goku, Goolmuddy, Gylgymesh, Hæl (aka Hael in this bot), Huen11, ItsMBSCastillo, iWushock, JDizzle, Jermoshua, JoeBu, John_Hatten2, josiah_4, kruste, Lqd, MaxBrand99, McSploosh, Melody (Choco), MrRager, Mochi, n1ghtmaree, Olivernoko, Orcaaa, PAINisGOD93, RoastyChicken, ShawnBond, Sizz, Smolder, Special_Delivery, Tarnym, Techno, Toad, Tolton, TooT, UnseenAxes, VesaN, Winter, WRLD_EATR, Youreprettycute, ZENICKS, and ZombieSlayer13

_Idea for Bot_: Trogburn

_Bot Coding_: BillyIdol ([Source Code on GitHub](https://github.com/deadloct/bitheroes-guide-bot))

_Initial Data Aggregation_: BillyIdol, ShawnBond, Trogdor, and ZombieSlayer13

_Honorable Mentions_: Robskino

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
