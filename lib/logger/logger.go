package logger

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func Debug(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Debug(args...)
}

func Debugf(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Debugf(format, args...)
}

func Info(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Info(args...)
}

func Infof(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Infof(format, args...)
}

func Warn(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Warn(args...)
}

func Warnf(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Warnf(format, args...)
}

func Error(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Error(args...)
}

func Errorf(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Errorf(format, args...)
}

func Panic(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Panic(args...)
}

func Panicf(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Panicf(format, args...)
}

func Fatal(i *discordgo.Interaction, args ...interface{}) {
	log.WithFields(getFields(i)).Fatal(args...)
}

func Fatalf(i *discordgo.Interaction, format string, args ...interface{}) {
	log.WithFields(getFields(i)).Fatalf(format, args...)
}

func getFields(i *discordgo.Interaction) log.Fields {
	return log.Fields{
		"channel_id":   i.ChannelID,
		"guild_id":     i.GuildID,
		"guild_locale": i.GuildLocale,
		"request_id":   i.ID,
	}
}
