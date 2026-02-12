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
	"sync"
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
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !guild.SDVerifySetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDVerify is not enabled")
			return nil
		}

		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
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
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}

		user := userOpt.UserValue(s)

		roleOpt, ok := optionMap["role"]
		if !ok {
			e := errors.New("role option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction, e)
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
				reportInteractionError(ctx, s, i.Interaction, e)
				return e
			}
		}
		for _, r := range roleToRemove {
			e := s.GuildMemberRoleRemove(i.GuildID, user.ID, r)
			if e != nil {
				logv2.Error(ctx, e, fmt.Sprintf("failed to remove role %s from user %s", r, user.ID))
				reportInteractionError(ctx, s, i.Interaction, e)
				return e
			}
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", fmt.Sprintf("Verification of user <@%s> with role <@&%s> is successful.\nThank you for using my service. Beep. Boop.\n", user.ID, religionRoleId)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
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
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}
		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
func (h *handler) sdQuestionOneCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !guild.SDQuestionOneSetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDQuestionOne is not enabled")
			return nil
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
			return nil
		}
		vqChannelId := guild.SDQuestionOneSetting.VettingQuestioningChannelId
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		userOpt, ok := optionMap["user"]
		if !ok {
			e := errors.New("user option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		user := userOpt.UserValue(s)

		err = s.GuildMemberRoleAdd(i.GuildID, user.ID, guild.Role.VettingQuestioning)
		if err != nil {
			logv2.Error(ctx, err, fmt.Sprintf("failed to add role %s to user %s", guild.Role.VettingQuestioning, user.ID))
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.ChannelMessageSendComplex(vqChannelId, &discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s>", user.ID),
			Embed: util.EmbedBuilder(
				"",
				fmt.Sprintf("Hey <@%s>! It looks like you missed question 1. Please re-read the <#%s> again, we assure you that the code is in there. Thank you!", user.ID, guild.Channel.RulesVetting),
			),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", fmt.Sprintf("Done. Please check <#%s>.", vqChannelId)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
func (h *handler) sdVettingQuestioningCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !guild.SDVettingQuestioningSetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDVettingQuestioning is not enabled")
			return nil
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
			return nil
		}
		vqChannelId := guild.SDVettingQuestioningSetting.VettingQuestioningChannelId
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		userOpt, ok := optionMap["user"]
		if !ok {
			e := errors.New("user option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		user := userOpt.UserValue(s)

		messageOpt, ok := optionMap["message"]
		if !ok {
			e := errors.New("message option is not found")
			logv2.Error(ctx, e, options)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		message := messageOpt.StringValue()

		err = s.GuildMemberRoleAdd(i.GuildID, user.ID, guild.Role.VettingQuestioning)
		if err != nil {
			logv2.Error(ctx, err, fmt.Sprintf("failed to add role %s to user %s", guild.Role.VettingQuestioning, user.ID))
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.ChannelMessageSendComplex(vqChannelId, &discordgo.MessageSend{
			Content: fmt.Sprintf("<@%s> %s", user.ID, message),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", fmt.Sprintf("Done. Please check <#%s>.", vqChannelId)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
func (h *handler) sdDetainCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !guild.SDQuestionOneSetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDQuestionOne is not enabled")
			return nil
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
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
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		user := userOpt.UserValue(s)
		member, err := s.GuildMember(i.GuildID, user.ID)
		if err != nil {
			return err
		}

		err = s.GuildMemberRoleAdd(i.GuildID, user.ID, guild.Role.Detained)
		if err != nil {
			logv2.Error(ctx, err, fmt.Sprintf("failed to add role %s to user %s", guild.Role.Detained, user.ID))
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		err = h.Provider.InsertRoles(ctx, i.GuildID, user.ID, member.Roles)
		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}
		wg := sync.WaitGroup{}
		for _, role := range member.Roles {
			wg.Add(1)
			go func(guildId, userId, role string) {
				defer wg.Done()
				err = s.GuildMemberRoleRemove(guildId, userId, role)
				if err != nil {
					logv2.Error(ctx, err, fmt.Sprintf("failed to remove role %s to user %s", role, userId))
				}
			}(i.GuildID, user.ID, role)
		}
		wg.Wait()

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", "Done. Please check #detained"),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}

func (h *handler) sdOfficeOfReadingsCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
			return nil
		}
		channelId := i.Interaction.ChannelID

		embeds, err := util.GenerateOfficeOfReadingsEmbeds()
		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.ChannelMessageSendEmbeds(channelId, embeds)
		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", ":white_check_mark:"),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
func (h *handler) sdCalendarCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
			return nil
		}
		channelId := i.Interaction.ChannelID

		embed, isMentionLatinCath, err := util.GenerateCalendarEmbed()
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}
		var msg string
		if isMentionLatinCath {
			msg = fmt.Sprintf("<@&%s>", guild.SDVerifySetting.ReligionRoleMap[domain.ReligionRoleKeyLatinCatholic])
		}
		_, err = s.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
			Content: msg,
			Embed:   embed,
		})
		if err != nil {
			logv2.Error(ctx, err)
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", ":white_check_mark:"),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
func (h *handler) sdUndetainCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		guild, ok := h.Config.GuildConfig[config.GuildId(i.GuildID)]
		if !ok {
			e := errors.New("guild is not found")
			logv2.Error(ctx, e, i)
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		if !guild.SDQuestionOneSetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "SDQuestionOne is not enabled")
			return nil
		}
		if !isMod(ctx, s, guild, i.Member.User.ID) {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("", fmt.Sprintf("You are not allowed to use this.")),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
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
			reportInteractionError(ctx, s, i.Interaction, e)
			return e
		}
		user := userOpt.UserValue(s)
		member, err := s.GuildMember(i.GuildID, user.ID)
		if err != nil {
			return err
		}

		err = s.GuildMemberRoleAdd(i.GuildID, user.ID, guild.Role.Detained)
		if err != nil {
			logv2.Error(ctx, err, fmt.Sprintf("failed to add role %s to user %s", guild.Role.Detained, user.ID))
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		err = h.Provider.InsertRoles(ctx, i.GuildID, user.ID, member.Roles)
		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}
		wg := sync.WaitGroup{}
		for _, role := range member.Roles {
			wg.Add(1)
			go func(guildId, userId, role string) {
				defer wg.Done()
				err = s.GuildMemberRoleRemove(guildId, userId, role)
				if err != nil {
					logv2.Error(ctx, err, fmt.Sprintf("failed to remove role %s to user %s", role, userId))
				}
			}(i.GuildID, user.ID, role)
		}
		wg.Wait()

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("", "Done. Please check #detained"),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction, err)
			return err
		}

		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
