package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

func (h *handler) sdVerifyCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction)
			return e
		}
		if !guild.SDVerifySetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDVerify is not enabled")
			return nil
		}
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		userOpt, ok := optionMap["user"]
		if !ok {
			e := errors.New("user option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction)
			return e
		}

		user := userOpt.UserValue(s)

		roleOpt, ok := optionMap["role"]
		if !ok {
			e := errors.New("role option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction)
			return e
		}

		roleType := roleOpt.StringValue()
		religionRoleId := guild.SDVerifySetting.ReligionRoleMap[domain.ReligionRoleKey(roleType)]
		roleToAdd := []string{guild.Role.ApprovedUser, religionRoleId}
		roleToRemove := []string{guild.Role.VettingQuestioning}
		for _, r := range guild.Role.Vetting {
			roleToRemove = append(roleToRemove, r)
		}
		for _, r := range roleToAdd {
			e := s.GuildMemberRoleAdd(i.GuildID, user.ID, r)
			if e != nil {
				logv2.Error(ctx, e, fmt.Sprintf("failed to add role %s to user %s", r, user.ID))
				reportInteractionError(ctx, s, i.Interaction)
				return e
			}
		}
		for _, r := range roleToRemove {
			e := s.GuildMemberRoleRemove(i.GuildID, user.ID, r)
			if e != nil {
				logv2.Error(ctx, e, fmt.Sprintf("failed to remove role %s from user %s", r, user.ID))
				reportInteractionError(ctx, s, i.Interaction)
				return e
			}
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", fmt.Sprintf("Verification of user <@%s> with role <@&%s> is successful.\nThank you for using my service. Beep. Boop.\n", user.ID, religionRoleId)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction)
			return err
		}

		welcomeMessageFormat := "Hey %s! %s just approved your vetting response. Welcome to the server. Feel free to tag us should you have further questions. Enjoy!"
		welcomeMessageEmbedFormat := "Welcome to Servus Dei, <@%s>! We are happy to have you! Make sure you check out <#%s> to gain access to the various channels we offer and please do visit <#%s> so you can understand our server better and take use of everything we have to offer. God Bless!"
		welcomeTitle := "Welcome to Servus Dei!"

		mod := i.Member
		content := fmt.Sprintf(welcomeMessageFormat, user.Mention(), mod.Mention())
		_, err = s.ChannelMessageSendComplex(guild.SDVerifySetting.WelcomeChannelId, &discordgo.MessageSend{
			Content: content,
			Embed: util.EmbedBuilder(
				welcomeTitle,
				fmt.Sprintf(welcomeMessageEmbedFormat, user.ID, guild.SDVerifySetting.ReactionRoleChannelId, guild.SDVerifySetting.ServerInformationChannelId),
				util.ImageUrl(util.RandomSDWelcomeImage()),
			),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction)
			return err
		}
		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
