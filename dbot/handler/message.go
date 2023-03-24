package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"regexp"
	"strings"
	"time"
)

func (h *handler) vettingQuestioningKeywordDetectionHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		cfg, ok := h.Config.GuildConfig[config.GuildId(m.GuildID)]
		if !ok {
			return
		}
		setting := cfg.DetectVettingQuestioningKeywordSetting
		channelId := setting.ChannelId
		if !setting.Enabled ||
			channelId != m.ChannelID {
			return
		}
		keyword := setting.Keyword
		if strings.Contains(util.ToAlphanumAndSpace(m.Content), keyword) {
			title := setting.Title
			rawDescription := setting.Description
			description := fmt.Sprintf(rawDescription, m.Author.ID)
			embed := util.EmbedBuilder(title, description)
			content := fmt.Sprintf("<@&%s>", cfg.Role.Moderator)
			_, err := s.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
				Content:   content,
				Embed:     embed,
				Reference: m.Reference(),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
			logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Keyword %s detected in guild %s. done replying acknowledgement message", keyword, m.GuildID))
			err = s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
			logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Done deleting message id %s", m.ID))
		}
	}
}

func detectVettingResponse(ctx context.Context, input string, setting config.InvalidVettingResponseSetting) bool {
	for _, flag := range setting.VettingResponseKeywordFlags {
		if strings.Contains(input, flag) {
			return true
		}
	}
	reg, err := regexp.Compile(".*2.*3.*4.*5.*6.*")
	if err != nil {
		logv2.Error(ctx, err)
		return false
	}
	return reg.MatchString(input)
}

func sanitizeVettingResponse(input string) string {
	var regex, err = regexp.Compile("(<(@|@&|#)(.*)?>)")
	if err != nil {
		return input
	}
	input = regex.ReplaceAllLiteralString(input, "")
	input = strings.ReplaceAll(util.ToAlphanum(input), "latinrite", "")
	return input
}

func isValidVettingResponse(ctx context.Context, input string, setting config.InvalidVettingResponseSetting) bool {
	input = sanitizeVettingResponse(input)
	if detectVettingResponse(ctx, input, setting) && !strings.Contains(input, setting.Keyword) {
		return false
	}
	return true
}

func (h *handler) invalidVettingResponseHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		cfg, ok := h.Config.GuildConfig[config.GuildId(m.GuildID)]
		if !ok {
			return
		}
		setting := cfg.InvalidVettingResponseSetting
		channelId := setting.ChannelId
		if !setting.Enabled ||
			channelId != m.ChannelID {
			return
		}

		if !isValidVettingResponse(ctx, m.Content, setting) {
			content := fmt.Sprintf("Hey <@%s>! It looks like you missed question 1. Please re-read the <#%s> again, we assure you that the code is in there. Thank you for your understanding.", m.Author.ID, cfg.Channel.RulesVetting)
			_, err := s.ChannelMessageSendEmbedReply(channelId, util.EmbedBuilder("", content), m.Reference())
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
		}

	}
}

func (h *handler) cmQuestionLimiterHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		cfg, ok := h.Config.GuildConfig[config.GuildId(m.GuildID)]
		if !ok {
			return
		}
		setting := cfg.CMQuestionLimiterSetting
		channelId := setting.ChannelId
		if !setting.Enabled ||
			channelId != m.ChannelID {
			return
		}
		if setting.UnlimitedRoleIds == nil {
			logv2.Error(ctx, fmt.Errorf("unlimited role ids is nil"))
			return
		}
		userId := m.Author.ID
		guildMember, err := s.GuildMember(m.GuildID, userId)
		if err != nil {
			logv2.Error(ctx, err)
			return
		}
		for _, r := range guildMember.Roles {
			if setting.UnlimitedRoleIds[r] {
				return
			}
		}

		//cek db apakah dia udah ada ask question in last 48 hours
		question, err := h.Provider.GetLatestQuestion(ctx, m.GuildID, userId)
		if err != nil {
			logv2.Error(ctx, err)
			return
		}

		if question.DocId != "" && question.Time.Add(time.Duration(setting.QuestionLimitDurationInMinutes)*time.Minute).After(time.Now()) {
			//send direct message to user
			channel, err := s.UserChannelCreate(userId)
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
			title := fmt.Sprintf("You just asked a question in the last %d hour(s). Please wait for a while before asking another question.", util.MinuteToHour(setting.QuestionLimitDurationInMinutes))
			_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{util.EmbedBuilder(title, m.Content)},
			})

			logv2.Debug(ctx, logv2.Info, fmt.Sprintf("User %s already asked question in last %d minutes. done replying acknowledgement message", userId, setting.QuestionLimitDurationInMinutes))
			err = s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logv2.Error(ctx, err)
				return
			}
			logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Done deleting message id %s", m.ID))
			return
		} else {
			go h.Provider.UpsertLatestQuestion(ctx, question)
		}
	}
}
