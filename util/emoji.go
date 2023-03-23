package util

import "github.com/bwmarrin/discordgo"

func IsSameEmoji(source string, target discordgo.Emoji) bool {
	return target.ID == source || target.Name == source
}
