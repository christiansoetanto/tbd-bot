package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

type commandHandler map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error
type componentHandler map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error

func (h *handler) GetCommandHandlers(ctx context.Context) ([]*discordgo.ApplicationCommand, commandHandler) {
	commandHandlers := commandHandler{
		domain.FeatureKeyPing:                 util.DecorateInteractionHandler(domain.FeatureKeyPing, h.pingCommandHandlerFunc(ctx)),
		domain.FeatureKeySDVerify:             util.DecorateInteractionHandler(domain.FeatureKeySDVerify, h.sdVerifyCommandHandlerFunc(ctx)),
		domain.FeatureKeySDQuestionOne:        util.DecorateInteractionHandler(domain.FeatureKeySDQuestionOne, h.sdQuestionOneCommandHandlerFunc(ctx)),
		domain.FeatureKeySDVettingQuestioning: util.DecorateInteractionHandler(domain.FeatureKeySDVettingQuestioning, h.sdVettingQuestioningCommandHandlerFunc(ctx)),
		domain.FeatureKeySDDetain:             util.DecorateInteractionHandler(domain.FeatureKeySDDetain, h.sdDetainCommandHandlerFunc(ctx)),
		domain.FeatureKeySDOfficeOfReadings:   util.DecorateInteractionHandler(domain.FeatureKeySDOfficeOfReadings, h.sdOfficeOfReadingsCommandHandlerFunc(ctx)),
		domain.FeatureKeySDCalendar:           util.DecorateInteractionHandler(domain.FeatureKeySDCalendar, h.sdCalendarCommandHandlerFunc(ctx)),
		domain.FeatureKeyCMQuestionOne:        util.DecorateInteractionHandler(domain.FeatureKeyCMQuestionOne, h.cmQuestionOneCommandHandlerFunc(ctx)),
		domain.FeatureKeyCMVerify:             util.DecorateInteractionHandler(domain.FeatureKeyCMVerify, h.cmVerifyCommandHandlerFunc(ctx)),
		domain.FeatureKeyCMPoll:               util.DecorateInteractionHandler(domain.FeatureKeyCMPoll, h.cmPollCommandHandlerFunc(ctx)),
		domain.FeatureKeyTSVerify:             util.DecorateInteractionHandler(domain.FeatureKeyTSVerify, h.tsVerifyCommandHandlerFunc(ctx)),
		domain.FeatureKeyTSQuestionOne:        util.DecorateInteractionHandler(domain.FeatureKeyTSQuestionOne, h.tsQuestionOneCommandHandlerFunc(ctx)),
		domain.FeatureKeyTSDetain:             util.DecorateInteractionHandler(domain.FeatureKeyTSDetain, h.tsDetainCommandHandlerFunc(ctx)),
		domain.FeatureKeyTSOfficeOfReadings:   util.DecorateInteractionHandler(domain.FeatureKeyTSOfficeOfReadings, h.tsOfficeOfReadingsCommandHandlerFunc(ctx)),
		domain.FeatureKeyTSCalendar:           util.DecorateInteractionHandler(domain.FeatureKeyTSCalendar, h.tsCalendarCommandHandlerFunc(ctx)),
	}
	applicationCommands := []*discordgo.ApplicationCommand{
		{
			Name:        domain.FeatureKeyPing,
			Description: "Pong",
		},
		{
			Name:        domain.FeatureKeySDVerify,
			Description: "verify user to Servus Dei server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "religion role to give",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Latin Catholic",
							Value: domain.ReligionRoleKeyLatinCatholic,
						},
						{
							Name:  "Eastern Catholic",
							Value: domain.ReligionRoleKeyEasternCatholic,
						},
						{
							Name:  "RCIA / Catechumen",
							Value: domain.ReligionRoleKeyRCIACatechumen,
						},
						{
							Name:  "Orthodox Christian",
							Value: domain.ReligionRoleKeyOrthodoxChristian,
						},
						{
							Name:  "Protestant",
							Value: domain.ReligionRoleKeyProtestant,
						},
						{
							Name:  "Non-Catholic",
							Value: domain.ReligionRoleKeyNonCatholic,
						},
						{
							Name:  "Atheist",
							Value: domain.ReligionRoleKeyAtheist,
						},
					},
				},
			},
		},
		{
			Name:        domain.FeatureKeyTSVerify,
			Description: "verify user to Terra Sancta server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "religion role to give",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Latin Catholic",
							Value: domain.ReligionRoleKeyLatinCatholic,
						},
						{
							Name:  "Eastern Catholic",
							Value: domain.ReligionRoleKeyEasternCatholic,
						},
						{
							Name:  "RCIA / Catechumen",
							Value: domain.ReligionRoleKeyRCIACatechumen,
						},
						{
							Name:  "Orthodox Christian",
							Value: domain.ReligionRoleKeyOrthodoxChristian,
						},
						{
							Name:  "Protestant",
							Value: domain.ReligionRoleKeyProtestant,
						},
						{
							Name:  "Non-Catholic",
							Value: domain.ReligionRoleKeyNonCatholic,
						},
						{
							Name:  "Atheist",
							Value: domain.ReligionRoleKeyAtheist,
						},
					},
				},
			},
		},
		{
			Name:        domain.FeatureKeyCMVerify,
			Description: "verify user to CapitalMindset server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeyCMPoll,
			Description: "poll tailored for CapitalMindset server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "question",
					Description: "The question you want to ask",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-1",
					Description: "The first option they can choose",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-2",
					Description: "The second option they can choose",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-3",
					Description: "The third option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-4",
					Description: "The fourth option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-5",
					Description: "The fifth option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-6",
					Description: "The sixth option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-7",
					Description: "The seventh option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-8",
					Description: "The eighth option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-9",
					Description: "The ninth option they can choose",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option-10",
					Description: "The tenth option they can choose",
					Required:    false,
				},
			},
		},
		{
			Name:        domain.FeatureKeySDVettingQuestioning,
			Description: "tag people in #vetting-questioning with a message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "message",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeySDDetain,
			Description: "detain",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeyTSDetain,
			Description: "detain",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to detain",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeySDOfficeOfReadings,
			Description: "send the 2nd reading of the Office of Readings",
		},
		{
			Name:        domain.FeatureKeyTSOfficeOfReadings,
			Description: "send the 2nd reading of the Office of Readings",
		},
		{
			Name:        domain.FeatureKeySDCalendar,
			Description: "send liturgical calendar for today",
		},
		{
			Name:        domain.FeatureKeyTSCalendar,
			Description: "send liturgical calendar for today",
		},
		{
			Name:        domain.FeatureKeySDQuestionOne,
			Description: "alert user that they missed question one answer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeyTSQuestionOne,
			Description: "alert user that they missed question one answer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
			},
		},
		{
			Name:        domain.FeatureKeyCMQuestionOne,
			Description: "alert user that they missed question one answer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "user to verify",
					Required:    true,
				},
			},
		},
	}
	return applicationCommands, commandHandlers

}

func (h *handler) buildCommandHandler(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			_, commandHandlers := h.GetCommandHandlers(ctx)
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				err := h(s, i)
				if err != nil {
					logv2.Error(ctx, err, fmt.Sprintf("Error while building command %v", i.ApplicationCommandData().Name))
				}
			}
		}
	}

}
func (h *handler) buildComponentHandler(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			handlers := componentHandler{
				domain.ComponentKeyCMVote:       util.DecorateInteractionHandler(domain.ComponentKeyCMVote, h.cmPollVoteHandlerFunc(ctx)),
				domain.ComponentKeyCMShowVoters: util.DecorateInteractionHandler(domain.ComponentKeyCMShowVoters, h.cmPollShowVotersHandlerFunc(ctx)),
			}
			if h, ok := handlers[i.MessageComponentData().CustomID]; ok {
				err := h(s, i)
				if err != nil {
					logv2.Error(ctx, err, fmt.Sprintf("Error while component %v", i.MessageComponentData().CustomID))
				}
			}
		}

	}

}
func (h *handler) pingCommandHandlerFunc(ctx context.Context) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		now := time.Now()
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
		msg := fmt.Sprintf("Pong! I responded in %dms", time.Since(now).Milliseconds())
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		if err != nil {
			logv2.Error(ctx, err)
			return err
		}
		logv2.Debug(ctx, logv2.Info, logv2.Finish)
		return nil
	}
}
