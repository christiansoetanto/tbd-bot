package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"strings"
)

type answersMap map[string][]answer
type answer struct {
	user *discordgo.User
	url  string
}

const (
	LimitPerRequest  = 100
	MaxMessageAmount = 3000
)

func (h *handler) questionMoverMessageReactionAddHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	ctx = logv2.InitRequestContext(ctx)
	ctx = logv2.InitFuncContext(ctx)
	return func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		guild, ok := h.Config.GuildConfig[config.GuildId(m.GuildID)]
		if !ok {
			return
		}
		setting := guild.QuestionDiscussionMoverSetting
		if !setting.Enabled || setting.QuestionChannelId != m.ChannelID || !util.IsSameEmoji(setting.TriggerReactionId, m.Emoji) || !isMod(ctx, s, guild, m.UserID) {
			return
		}
		questionId := m.MessageID
		message, err := s.ChannelMessage(setting.QuestionChannelId, questionId)
		if err != nil {
			logv2.Error(ctx, err)
			return
		}

		question, questionAskerId := message.Content, message.Author.ID
		answers, err := getAllAnswers(ctx, s, m.ChannelID, m.MessageID, setting)
		if err != nil {
			return
		}
		if answers == nil || len(answers) == 0 {
			msgLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guild.GuildId, setting.QuestionChannelId, message.ID)
			logv2.Debug(ctx, logv2.Info, fmt.Sprintf("No answers found for question: %s", msgLink))
			return
		}
		answers, err = generateAnswerUrl(ctx, s, answers, question, guild.GuildId)
		if err != nil {
			return
		}
		err = archiveQuestion(ctx, s, answers, questionAskerId, questionId, question, setting)
		if err != nil {
			return
		}
	}
}

func getAllAnswers(ctx context.Context, s *discordgo.Session, channelId, messageId string, setting config.QuestionDiscussionMoverSetting) (answersMap, error) {
	a := make(answersMap)
	for _, stg := range setting.DiscussionAnswerSetting {
		cur, err := getAnswersForChannel(ctx, s, channelId, messageId, stg.ReactionId)
		if err != nil {
			logv2.Error(ctx, err)
			return nil, err
		}
		if cur != nil {
			a[stg.ChannelId] = cur
		}
	}
	return a, nil
}

func getAnswersForChannel(ctx context.Context, s *discordgo.Session, channelId, messageId, reactionId string) ([]answer, error) {
	ctx = logv2.InitFuncContext(ctx)
	users, err := s.MessageReactions(channelId, messageId, reactionId, 0, "", "")
	if err != nil {
		logv2.Error(ctx, err)
		return nil, err
	}
	answers := answersBuilder(users)
	return answers, err
}

func answersBuilder(users []*discordgo.User) []answer {
	var answers []answer
	for _, user := range users {
		answers = append(answers, answer{
			user: user,
		})
	}
	return answers
}

func generateAnswerUrl(ctx context.Context, s *discordgo.Session, answerMap answersMap, question string, guildId config.GuildId) (answersMap, error) {
	ctx = logv2.InitFuncContext(ctx)
	for channelId := range answerMap {
		answers := answerMap[channelId]
		lastMessageId := ""
		totalAnswerToBeFound := len(answers)
		for i := 0; i < MaxMessageAmount/LimitPerRequest && totalAnswerToBeFound > 0; i++ {
			messages, err := s.ChannelMessages(channelId, LimitPerRequest, lastMessageId, "", "")
			if err != nil {
				logv2.Error(ctx, err)
				return nil, err
			}
			if len(messages) == 0 {
				return answerMap, nil
			}
			lastMessageId = messages[len(messages)-1].ID
			//loop setiap message, cari apakah message itu dimiliki oleh user yang menjawab
			for _, message := range messages {
				for i := range answers {
					if message.Author.ID == answers[i].user.ID {
						//TODO ganti ke levenshtein
						if strings.Contains(util.ToAlphanum(ctx, message.Content), util.ToAlphanum(ctx, question)) {
							answerLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildId, channelId, message.ID)
							answers[i].url = answerLink
							totalAnswerToBeFound -= 1
						}
					}
				}
				if totalAnswerToBeFound <= 0 {
					break
				}
			}
		}
	}
	return answerMap, nil
}

func archiveQuestion(ctx context.Context, s *discordgo.Session, answers answersMap, questionAskerId, questionId, questionContent string, setting config.QuestionDiscussionMoverSetting) error {
	ctx = logv2.InitFuncContext(ctx)
	title := ""
	description := fmt.Sprintf("\nQuestion by <@%s>:\n\n%s\n", questionAskerId, questionContent)
	fieldValue := ""
	fieldName := "Answer(s):"
	for _, answers := range answers {
		for _, answer := range answers {
			fieldValue += fmt.Sprintf("\n- <@%s>", answer.user.ID)
			if len(answer.url) > 0 {
				fieldValue += fmt.Sprintf(" [jump to answer!](%s)", answer.url)
			}
		}
	}

	var fields []*discordgo.MessageEmbedField
	if fieldValue != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fieldName,
			Value:  fieldValue,
			Inline: false,
		})
	}
	embed := util.EmbedBuilder(title, description, fields)

	_, err := s.ChannelMessageSendEmbed(setting.AnsweredQuestionChannelId, embed)

	if err != nil {
		logv2.Error(ctx, err)
		return err
	}

	err = s.ChannelMessageDelete(setting.QuestionChannelId, questionId)
	if err != nil {
		logv2.Error(ctx, err)
		return err
	}

	return nil
}
