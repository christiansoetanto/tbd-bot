package util

import "github.com/bwmarrin/discordgo"

type ImageUrl string

const (
	LogoURL           = "https://cdn.discordapp.com/avatars/767426889294938112/0e100e9fec18866892ed0c875b341926.png"
	FooterText        = "2023 | Made by soetanto"
	GoldenYellowColor = 16769280
)

func EmbedBuilder(title string, description string, param ...interface{}) *discordgo.MessageEmbed {
	var imageUrl string
	var fields []*discordgo.MessageEmbedField
	for _, p := range param {
		switch v := p.(type) {
		case ImageUrl:
			imageUrl = string(v)
		case []*discordgo.MessageEmbedField:
			fields = v
		}

	}
	embed := &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       title,
		Description: description,
		Color:       GoldenYellowColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    FooterText,
			IconURL: LogoURL,
		},
		Fields: fields,
	}
	if imageUrl != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: imageUrl,
		}
	}
	return embed
}
