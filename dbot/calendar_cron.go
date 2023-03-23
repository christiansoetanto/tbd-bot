package dbot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AllLiturgicalDays struct {
	LiturgicalDaysEn []LiturgicalDay
	LiturgicalDaysLa []LiturgicalDay
}
type LiturgicalDay struct {
	Key                   string        `json:"key"`
	Date                  string        `json:"date"`
	Precedence            string        `json:"precedence"`
	Rank                  string        `json:"rank"`
	IsHolyDayOfObligation bool          `json:"isHolyDayOfObligation"`
	IsOptional            bool          `json:"isOptional"`
	Martyrology           []interface{} `json:"martyrology"`
	Titles                []string      `json:"titles"`
	Calendar              Calendar      `json:"calendar"`
	Cycles                Cycles        `json:"cycles"`
	Name                  string        `json:"name"`
	RankName              string        `json:"rankName"`
	ColorName             []string      `json:"colorName"`
	SeasonNames           []string      `json:"seasonNames"`
}
type Calendar struct {
	WeekOfSeason          int    `json:"weekOfSeason,omitempty"`
	DayOfSeason           int    `json:"dayOfSeason,omitempty"`
	DayOfWeek             int    `json:"dayOfWeek,omitempty"`
	NthDayOfWeekInMonth   int    `json:"nthDayOfWeekInMonth,omitempty"`
	StartOfSeason         string `json:"startOfSeason,omitempty"`
	EndOfSeason           string `json:"endOfSeason,omitempty"`
	StartOfLiturgicalYear string `json:"startOfLiturgicalYear,omitempty"`
	EndOfLiturgicalYear   string `json:"endOfLiturgicalYear,omitempty"`
}
type Cycles struct {
	ProperCycle  string `json:"properCycle"`
	SundayCycle  string `json:"sundayCycle"`
	WeekdayCycle string `json:"weekdayCycle"`
	PsalterWeek  string `json:"psalterWeek"`
}
type Martyrology struct {
	Key               string   `json:"key"`
	CanonizationLevel string   `json:"canonizationLevel"`
	DateOfDeath       int      `json:"dateOfDeath"`
	Titles            []string `json:"titles,omitempty"`
}

type Messages struct {
	Messages []MessageItem `json:"messages"`
}
type MessageItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (u *usecase) generateCalendarEmbed(ctx context.Context) (*discordgo.MessageEmbed, error) {
	functionsUrl := os.Getenv("ROMCAL_API_FUNCTIONS_URL")
	response, err := http.Get(functionsUrl)
	if err != nil {
		logv2.Error(ctx, err)
		return nil, err
	}
	data, _ := io.ReadAll(response.Body)

	var allLiturgicalDays AllLiturgicalDays
	err = json.Unmarshal(data, &allLiturgicalDays)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	text, isAnyHDO := getCelebrations(allLiturgicalDays.LiturgicalDaysEn)
	title := fmt.Sprintf("%s, %d %s %d", currentTime.Weekday(), currentTime.Day(), currentTime.Month(), currentTime.Year())
	var fields []*discordgo.MessageEmbedField
	if isAnyHDO && currentTime.Weekday() != time.Sunday {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Disclaimer:",
			Value:  "The rule for Holy Day of Obligation can vary between dioceses as per Code of Canon Law Can. 1246. Please consult with your local priest if you have any doubt.",
			Inline: false,
		})
	}
	embed := util.EmbedBuilder(title, text, fields)
	return embed, nil
}

func (u *usecase) liturgicalCalendarCronJob(ctx context.Context) func() {
	return func() {
		embed, err := u.generateCalendarEmbed(ctx)
		if err != nil {
			logv2.Error(ctx, err)
			return
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyCalendarCron] {
				_, err = u.Session.ChannelMessageSendEmbed(config.Channel.LiturgicalCalendarDiscussion, embed)
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}

func getCelebrations(liturgicalDays []LiturgicalDay) (string, bool) {
	text := "The Roman Catholic Church is celebrating:\n"
	isAnyHDO := false
	for _, day := range liturgicalDays {
		text += "• "
		//[day, date] //if memorial/feast/solemnity [rank] [name] in [seasonName] season.
		rank, rankName, isHolyDayOfObligation, name, seasonNames := strings.ToLower(day.Rank), day.RankName,
			day.IsHolyDayOfObligation, day.Name, day.SeasonNames
		if rank == "memorial" || rank == "feast" || rank == "solemnity" {
			text += fmt.Sprintf("%s of %s", cases.Title(language.AmericanEnglish).String(rankName), name)
			if len(seasonNames) > 0 {
				text += fmt.Sprintf(" in the %s", seasonNames[0])
			}
		} else {
			text += name
		}

		if isHolyDayOfObligation {
			text += ", a Holy Day of Obligation"
			isAnyHDO = true
		}

		text += ".\n"
	}
	return text, isAnyHDO

}
