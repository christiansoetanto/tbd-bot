package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"time"
)

func reportInteractionError(ctx context.Context, s *discordgo.Session, i *discordgo.Interaction) {
	ctx = logv2.InitFuncContext(ctx)
	reqId := logv2.GetCtxReqId(ctx)
	timeNow := time.Now().Format(time.RFC3339)
	type reporter struct {
		ReqId string `json:"req_id"`
		Time  string `json:"time"`
	}
	r := reporter{
		ReqId: reqId,
		Time:  timeNow,
	}

	msg, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		logv2.Error(ctx, err, r)
		return
	}

	_, err = s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Embeds: util.EmbedsBuilder("", "Sorry, an error occured. Please try again or if problem persists, contact the developer with below data.", []*discordgo.MessageEmbedField{
			{
				Name:   "",
				Value:  fmt.Sprintf("```%s```", string(msg)),
				Inline: false,
			},
		}),
	})
	if err != nil {
		logv2.Error(ctx, err)
		return
	}
	return
}

func isMod(ctx context.Context, s *discordgo.Session, guildCfg config.GuildConfig, userId string) bool {
	ctx = logv2.InitFuncContext(ctx)
	member, err := s.GuildMember(string(guildCfg.GuildId), userId)
	if err != nil {
		logv2.Error(ctx, err)
		return false
	}

	for _, userRole := range member.Roles {
		for _, allowedRole := range guildCfg.Role.ModeratorOrAbove {
			if userRole == allowedRole {
				return true
			}
		}
	}

	return false
}

func isCMGoldMember(ctx context.Context, s *discordgo.Session, guildCfg config.GuildConfig, userId string) bool {
	ctx = logv2.InitFuncContext(ctx)
	member, err := s.GuildMember(string(guildCfg.GuildId), userId)
	if err != nil {
		logv2.Error(ctx, err)
		return false
	}

	for _, userRole := range member.Roles {
		for _, allowedRole := range guildCfg.Role.CMGoldMember {
			if userRole == allowedRole {
				return true
			}
		}
	}

	return false
}

func buildPollUI(poll domain.Poll) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField

	var val string
	for _, option := range poll.Options {
		val += fmt.Sprintf("%d votes - %s\n", option.Weight, option.Value)
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Options",
		Value: val,
	})
	return util.EmbedBuilder("Poll", poll.Question, fields)
}
func buildPollShowVotersUI(ctx context.Context, s *discordgo.Session, guildCfg config.GuildConfig, poll domain.Poll) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField
	for _, option := range poll.Options {
		var val string
		for _, voter := range option.Voters {
			val += fmt.Sprintf("<@%s>", voter.UserId)
			if isCMGoldMember(ctx, s, guildCfg, voter.UserId) {
				val += fmt.Sprintf(" %s", guildCfg.Reaction.Gold)
			}
			val += fmt.Sprintf("\n")
		}
		if val == "" {
			val = "No one voted yet"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  option.Value,
			Value: val,
		})
	}
	return util.EmbedBuilder("Voters", poll.Question, fields)
}
