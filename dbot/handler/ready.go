package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/logv2"
)

func (h *handler) readyHandler(ctx context.Context) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Logged in as: %v#%v id: %v", s.State.User.Username, s.State.User.Discriminator, s.State.User.ID))
	}
}
