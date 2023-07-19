package dbot

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (u *usecase) fridayCronJob(ctx context.Context) func() {
	return func() {
		c := util.RandomSDFridayMemeImage()
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyFridayCron] {
				_, err := u.Session.ChannelMessageSend(config.Channel.ChristianMemes, c)
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}
