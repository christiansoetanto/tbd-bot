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

func (h *handler) cmPollVoteHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		pollId := i.Message.ID
		channelId := i.ChannelID
		voteValue := i.Interaction.MessageComponentData().Values
		userId := i.Interaction.Member.User.ID
		guildId := i.GuildID
		guildCfg := h.Config.GuildConfig[config.GuildId(guildId)]
		if len(voteValue) == 0 {
			err := errors.New("vote value is empty")
			logv2.Error(ctx, err, i.Interaction.MessageComponentData())
			return err
		}
		newAnswer := voteValue[0]
		poll, err := h.Provider.GetPoll(ctx, pollId)
		if err != nil {
			return err
		}

		logv2.Debug(ctx, logv2.Info, poll, "Before update")

		oldAnswer := ""
		for _, option := range poll.Options {
			for _, voter := range option.Voters {
				if voter.UserId == userId {
					oldAnswer = option.Value
				}
			}
		}

		//check if user is gold
		userWeight := 1
		if isCMGoldMember(ctx, s, guildCfg, userId) {
			userWeight = 4
		}
		//if user is gold,weight = 4

		//build new poll
		for j := 0; j < len(poll.Options); j++ {
			if oldAnswer != "" && poll.Options[j].Value == oldAnswer {
				//remove vote
				for k := 0; k < len(poll.Options[k].Voters); k++ {
					if poll.Options[k].Voters[k].UserId == userId {
						poll.Options[k].Voters = util.VotersRemoveIndex(poll.Options[k].Voters, k)
					}
				}
				poll.Options[j].Weight -= userWeight //also remove weight

				//also remove weight
			}
			if poll.Options[j].Value == newAnswer {
				//add vote
				poll.Options[j].Voters = append(poll.Options[j].Voters, domain.Voter{
					UserId: userId,
				})

				//also add weight
				poll.Options[j].Weight += userWeight
			}
		}

		err = h.Provider.UpsertPoll(ctx, poll)
		if err != nil {
			return err
		}
		logv2.Debug(ctx, logv2.Info, poll, "After update")

		//get old message
		oldMessage, err := s.ChannelMessage(channelId, pollId)
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}
		_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Components: &oldMessage.Components,
			ID:         pollId,
			Channel:    channelId,
			Embed:      buildPollUI(poll),
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{util.EmbedBuilder("Poll", fmt.Sprintf("Success vote for %s", newAnswer))},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}

		return nil
	}
}

func (h *handler) cmPollShowVotersHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		pollId := i.Message.ID
		userId := i.Interaction.Member.User.ID
		guildId := i.GuildID
		guildCfg := h.Config.GuildConfig[config.GuildId(guildId)]

		if !isMod(ctx, s, guildCfg, userId) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{util.EmbedBuilder("Poll", "Sorry, you are not allowed to use this feature. Please ask the mods for further assistance.")},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logv2.Error(ctx, err)
				return err
			}
			return nil
		}

		poll, err := h.Provider.GetPoll(ctx, pollId)
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{buildPollShowVotersUI(ctx, s, guildCfg, poll)},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}
		return nil
	}
}
