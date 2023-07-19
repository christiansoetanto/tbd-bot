package dbot

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (u *usecase) officeOfReadingsCronJob(ctx context.Context) func() {
	return func() {
		embed, err := util.GenerateOfficeOfReadingsEmbeds()
		if err != nil {
			return
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyOfficeOfReadingsCron] {
				_, err = u.Session.ChannelMessageSendEmbeds(config.Channel.OfficeOfReadings, embed)
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}
