package lib

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const TickerInterval = 10 * time.Second

type GuildIndex struct {
	guildNames map[string]string
	sess       *discordgo.Session
	stop       chan struct{}
	ticker     *time.Ticker

	sync.Mutex
}

func NewGuildIndex(sess *discordgo.Session) *GuildIndex {
	return &GuildIndex{
		guildNames: make(map[string]string),
		sess:       sess,
		stop:       make(chan struct{}),
		ticker:     time.NewTicker(TickerInterval),
	}
}

func (g *GuildIndex) Start() {
	g.updateGuilds()

	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.updateGuilds()
			case <-g.stop:
				g.ticker.Stop()
				return
			}
		}
	}()
}

func (g *GuildIndex) Stop() {
	g.ticker.Stop()
	close(g.stop)
}

func (g *GuildIndex) GetGuildName(id string) string {
	g.Lock()
	defer g.Unlock()

	name, ok := g.guildNames[id]
	if !ok {
		return ""
	}

	return name
}

func (g *GuildIndex) updateGuilds() {
	g.Lock()
	defer g.Unlock()

	for _, guild := range g.sess.State.Guilds {
		if guild != nil {
			// Name is probably only available if access was granted by admin
			g.guildNames[guild.ID] = guild.Name
		}
	}
}
