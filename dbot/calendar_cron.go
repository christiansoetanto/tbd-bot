package dbot

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (u *usecase) liturgicalCalendarCronJob(ctx context.Context) func() {
	return func() {
		embed, isMentionLatinCath, err := util.GenerateCalendarEmbed()
		if err != nil {
			logv2.Error(ctx, err)
			return
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyCalendarCron] {
				var msg string
				if isMentionLatinCath {
					msg = fmt.Sprintf("<@&%s>", config.SDVerifySetting.ReligionRoleMap[domain.ReligionRoleKeyLatinCatholic])
				}
				_, err = u.Session.ChannelMessageSendComplex(config.Channel.LiturgicalCalendarDiscussion, &discordgo.MessageSend{
					Content: msg,
					Embed:   embed,
				})
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}
