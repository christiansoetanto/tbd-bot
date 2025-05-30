package util

import (
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"time"
)

type ImageUrl string

const (
	//LogoURL           = "https://cdn.discordapp.com/avatars/974311059680821268/02318b836ae4e2106d383c88a6f7370e.png"
	LogoURL           = "https://cdn.discordapp.com/avatars/767426889294938112/0e100e9fec18866892ed0c875b341926.png"
	FooterText        = "2023 | Made by soetanto"
	GoldenYellowColor = 16769280
)

var sdCounter = 0
var tsCounter = 0
var cmCounter = 0

func RandomSDWelcomeImage() string {
	welcomeImageURL := "https://cdn.discordapp.com/attachments/751174152588623912/976921809607880714/You_Doodle_2022-05-19T18_58_15Z.jpg"
	welcomeImage2URL := "https://media.discordapp.net/attachments/751174152588623912/975368929008558130/Screenshot_2022-05-11_at_11.42.51_PM.png"
	imgs := []string{welcomeImageURL, welcomeImage2URL}
	if sdCounter >= len(imgs) {
		sdCounter = 0
	}
	img := imgs[sdCounter]
	sdCounter++
	return img
}
func RandomTSWelcomeImage() string {
	welcomeImageURL := "https://cdn.discordapp.com/attachments/888111411052552243/1165307867004416111/83clat.jpg?ex=65466085&is=6533eb85&hm=f8f7bd9ef0f3043fbbf773a04e1e93f31d01a985bcb5405df4dc976db4b1ef0c&"
	welcomeImage2URL := "https://cdn.discordapp.com/attachments/888111411052552243/1165307867394494594/83cl4c.jpg?ex=65466085&is=6533eb85&hm=204f31fb46ecc12eae1faf58de92d00cbc607d0e981fe01a579a2c56a8e8e3c2&"
	imgs := []string{welcomeImageURL, welcomeImage2URL}
	if tsCounter >= len(imgs) {
		tsCounter = 0
	}
	img := imgs[tsCounter]
	tsCounter++
	return img
}
func RandomSDFridayMemeImage() string {
	imgs := []string{
		"https://imgur.com/wq7TBX0",
		"https://imgur.com/l0y0w9c",
		"https://imgur.com/wkwVjjI",
		"https://imgur.com/CIeRgm9",
		"https://imgur.com/7h7iuvK",
	}
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(imgs))
	img := imgs[randomIndex]
	return img
}

func RandomCMWelcomeImage() string {
	welcomeImageURL := "https://cdn.discordapp.com/attachments/751174152588623912/976921809607880714/You_Doodle_2022-05-19T18_58_15Z.jpg"
	welcomeImage2URL := "https://media.discordapp.net/attachments/751174152588623912/975368929008558130/Screenshot_2022-05-11_at_11.42.51_PM.png"
	imgs := []string{welcomeImageURL, welcomeImage2URL}
	if cmCounter >= len(imgs) {
		cmCounter = 0
	}
	img := imgs[cmCounter]
	cmCounter++
	return img
}

func EmbedsBuilder(title string, description string, param ...interface{}) *[]*discordgo.MessageEmbed {
	return &[]*discordgo.MessageEmbed{EmbedBuilder(title, description, param...)}
}
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
