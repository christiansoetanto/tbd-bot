package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"strings"
)

func (h *handler) keywordDetectionHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		if strings.Contains(util.ToAlphanumAndSpace(ctx, m.Content), keyword) {
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
