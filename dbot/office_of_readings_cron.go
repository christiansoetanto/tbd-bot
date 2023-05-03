package dbot

import (
	"context"
	"errors"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"net/http"
	"strings"
	"time"
)

func (u *usecase) generateOfficeOfReadingsEmbed(ctx context.Context) (*discordgo.MessageEmbed, error) {
	res, err := http.Get("https://ibreviary.com/m2/breviario.php?s=ufficio_delle_letture")
	if err != nil {
		logv2.Error(ctx, err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logv2.Error(ctx, err)
		return nil, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		logv2.Error(ctx, err)
		return nil, err
	}
	var title, text string
	title = fmt.Sprintf("Office of Readings for %s", time.Now().Format("Monday, 02 January 2006"))
	doc.Find(".rubrica").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		if s.Text() == "SECOND READING" {
			parent := s.Parent()
			converter := md.NewConverter("", true, nil)
			markdown := converter.Convert(parent)
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
			//cut Responsory
			responsorySplitted := strings.Split(markdown, "RESPONSORY")
			text = responsorySplitted[0]
			secondReadingSplitted := strings.Split(text, "SECOND READING")
			text = secondReadingSplitted[1]
		}
	})

	if text == "" {
		return nil, errors.New("no second reading found")
	}

	embed := util.EmbedBuilder(title, text)
	return embed, nil
}

func (u *usecase) officeOfReadingsCronJob(ctx context.Context) func() {
	return func() {
		embed, err := u.generateOfficeOfReadingsEmbed(ctx)
		if err != nil {
			return
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyOfficeOfReadingsCron] {
				_, err = u.Session.ChannelMessageSendEmbed(config.Channel.OfficeOfReadings, embed)
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}
