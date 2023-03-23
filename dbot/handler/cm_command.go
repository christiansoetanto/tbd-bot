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

func (h *handler) cmVerifyCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
		if !guild.CMVerifySetting.Enabled {
			logv2.Debug(ctx, logv2.Info, "CMVerify is not enabled")
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

		roleToAdd := []string{guild.Role.ApprovedUser}
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
			Embeds: util.EmbedsBuilder("", fmt.Sprintf("Verification of user <@%s> is successful.\nThank you for using my service. Beep. Boop.\n", user.ID)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction)
			return err
		}

		welcomeMessageEmbedFormat := "Welcome to Capital Mindset!, <@%s>! We are happy to have you! Make sure you check out <#%s> to gain access to the various channels we offer. God Bless!"
		welcomeTitle := "Welcome to Capital mindset!"
		welcomeMessageFormat := "Hey %s! %s just approved your vetting response. Welcome to the server. Feel free to tag us should you have further questions. Enjoy!"

		mod := i.Member
		content := fmt.Sprintf(welcomeMessageFormat, user.Mention(), mod.Mention())
		_, err = s.ChannelMessageSendComplex(guild.CMVerifySetting.WelcomeChannelId, &discordgo.MessageSend{
			Content: content,
			Embed: util.EmbedBuilder(
				welcomeTitle,
				fmt.Sprintf(welcomeMessageEmbedFormat, user.ID, guild.CMVerifySetting.ReactionRoleChannelId),
				util.ImageUrl(util.RandomCMWelcomeImage()),
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
func (h *handler) cmPollCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
		guildId := i.GuildID
		channelId := i.ChannelID
		//build option here
		options := i.ApplicationCommandData().Options
		questionValue, pollOptions := parseOptions(options)

		if util.HasDuplicateItem(pollOptions) {
			//return error
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: util.EmbedsBuilder("Error", "Duplicate option is not allowed"),
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
		}

		domainPollOptions := buildDomainPollOptions(pollOptions)
		selectMenuOptions := buildSelectMenuOptions(pollOptions)

		pollMessage, err := s.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
			Embed: util.EmbedBuilder("Poll", "Poll is being created. Please wait..."),
		})
		pollId := pollMessage.ID
		pollMessageLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildId, channelId, pollId)
		poll := domain.Poll{
			Id:       pollId,
			Question: questionValue,
			Options:  domainPollOptions,
		}
		err = h.Provider.UpsertPoll(ctx, poll)
		if err != nil {
			logv2.Error(ctx, err)
			e := s.ChannelMessageDelete(channelId, pollId)
			if e != nil {
				logv2.Error(ctx, e)
			}
			return err
		}
		_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							MenuType:    discordgo.StringSelectMenu,
							CustomID:    "vote",
							Placeholder: "Choose an option to vote for.",
							Options:     selectMenuOptions,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Show Voters",
							Style:    discordgo.PrimaryButton,
							CustomID: domain.ComponentKeyCMShowVoters,
						},
					},
				},
			},
			ID:      pollId,
			Channel: channelId,
			Embed:   buildPollUI(poll),
		})

		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: util.EmbedsBuilder("Poll created", fmt.Sprintf("Click this link to see your poll!\n %s\n", pollMessageLink)),
		})

		if err != nil {
			logv2.Error(ctx, err)
			reportInteractionError(ctx, s, i.Interaction)
			return err
		}

		return nil

	}
}

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (question string, pollOptions []string) {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	question = optionMap["question"].StringValue()
	for j := 1; j <= 10; j++ {
		option := optionMap[fmt.Sprintf("option-%d", j)]
		if option != nil {
			pollOptions = append(pollOptions, option.StringValue())
		}
	}
	return
}
func buildDomainPollOptions(options []string) []domain.Option {
	var pollOptions []domain.Option
	for _, o := range options {
		pollOptions = append(pollOptions, domain.Option{
			Value: o,
		})
	}
	return pollOptions
}

func buildSelectMenuOptions(options []string) []discordgo.SelectMenuOption {
	var selectMenuOptions []discordgo.SelectMenuOption
	for _, o := range options {
		selectMenuOptions = append(selectMenuOptions, discordgo.SelectMenuOption{
			Label: o,
			Value: o,
		})
	}
	return selectMenuOptions
}
