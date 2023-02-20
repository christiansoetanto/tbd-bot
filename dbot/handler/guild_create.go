package handler

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/logv2"
)

func (h *handler) guildCreateHandler(ctx context.Context) func(s *discordgo.Session, gc *discordgo.GuildCreate) {
	return func(s *discordgo.Session, gc *discordgo.GuildCreate) {
		logv2.Debug(ctx, logv2.Info, "Loaded guild id %v, guild name: %v", gc.Guild.ID, gc.Guild.Name)
	}
}
