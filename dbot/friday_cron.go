package dbot

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
)

func (u *usecase) fridayCronJob(ctx context.Context) func() {
	return func() {
		content := []string{
			"https://scontent.fcgk33-1.fna.fbcdn.net/v/t39.30808-6/275562959_2318020378347691_1790417404198505656_n.png?_nc_cat=103&ccb=1-7&_nc_sid=730e14&_nc_eui2=AeF0tRZVTolXWn4gZjfQnCDI_d9Jq_-iytr930mr_6LK2nDFkG4f2ipaedFv99j8Hw_PmU7OsaFC51aS7c364mX7&_nc_ohc=LGVZi3DVMNEAX9pseoP&_nc_ht=scontent.fcgk33-1.fna&oh=00_AfBOgkuKoV4QhWKNKToExDmtIj2fjIv0ZylHmw6IlvG15g&oe=64B56610",
			"https://pbs.twimg.com/media/FM5fbOsWUAUTTLd?format=jpg&name=small",
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyFridayCron] {
				for _, c := range content {
					_, err := u.Session.ChannelMessageSend(config.Channel.ChristianMemes, c)
					if err != nil {
						logv2.Error(ctx, err)
					}
				}
			}
		}
	}
}
