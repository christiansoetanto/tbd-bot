package handler

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/provider"
	"github.com/christiansoetanto/tbd-bot/util"
)

type handler struct {
	Config   config.Config
	Provider provider.Provider
}

type Resource struct {
	Config   config.Config
	Provider provider.Provider
}

type Handler interface {
	GetHandlers(ctx context.Context) []interface{}
	GetCommandHandlers(ctx context.Context) ([]*discordgo.ApplicationCommand, commandHandler)
}

var obj Handler
var once sync.Once

// GetProvider get provider client
func GetHandler(resource *Resource) Handler {
	once.Do(func() {
		obj = &handler{
			Config:   resource.Config,
			Provider: resource.Provider,
		}
	})
	return obj
}

func (h *handler) GetHandlers(ctx context.Context) []interface{} {
	return []interface{}{
		util.DecorateEventHandler("ready", h.readyHandler(ctx)),
		util.DecorateEventHandler("guild_create", h.guildCreateHandler(ctx)),
		util.DecorateEventHandler("vetting_questioning_keyword", h.vettingQuestioningKeywordDetectionHandler(ctx)),
		util.DecorateEventHandler("build_command", h.buildCommandHandler(ctx)),
		util.DecorateEventHandler("build_component", h.buildComponentHandler(ctx)),
		util.DecorateEventHandler("question_mover_reaction_add", h.questionMoverMessageReactionAddHandler(ctx)),
		util.DecorateEventHandler("cm_question_limiter", h.cmQuestionLimiterHandler(ctx)),
		util.DecorateEventHandler("invalid_vetting_response", h.invalidVettingResponseHandler(ctx)),
	}
}
