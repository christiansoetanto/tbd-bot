package handler

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/provider"
	"sync"
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
		h.readyHandler(ctx),
		h.guildCreateHandler(ctx),
		h.vettingQuestioningKeywordDetectionHandler(ctx),
		h.buildCommandHandler(ctx),
		h.buildComponentHandler(ctx),
		h.questionMoverMessageReactionAddHandler(ctx),
		h.cmQuestionLimiterHandler(ctx),
		h.invalidVettingResponseHandler(ctx),
	}
}
