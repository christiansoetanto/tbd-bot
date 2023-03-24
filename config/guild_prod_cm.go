package config

import "github.com/christiansoetanto/tbd-bot/domain"

func getCapitalMindsetProdGuildConfig() GuildConfig {
	return GuildConfig{
		GuildId:   "1077804660997500980",
		GuildName: "dev - capital mindset",
		Channel: Channel{
			RulesVetting: "1088360872558211164",
		},
		Role: Role{
			Vetting:            []string{"1078236690004594759", "1078236722405593118"},
			VettingQuestioning: "1078241164681019462",
			ApprovedUser:       "1078236749861494807",
			Moderator:          "1078336581703843840",
			ModeratorOrAbove:   []string{"1078336581703843840", "1078336615958708384"},
			CMGoldMember:       []string{"1078236771751563305", "1078236793402576907"},
		},
		Reaction: Reaction{
			Gold: "<:gold:1088346927801843752>",
		},
		RegisteredFeature: map[string]bool{
			domain.FeatureKeyPing:          true,
			domain.FeatureKeyCMVerify:      true,
			domain.FeatureKeyCMPoll:        true,
			domain.FeatureKeyCMQuestionOne: true,
		},
		DetectVettingQuestioningKeywordSetting: DetectVettingQuestioningKeywordSetting{
			Enabled:     true,
			Keyword:     "stong",
			ChannelId:   "1078237386615562253",
			Title:       "title",
			Description: "description",
		},
		CMVerifySetting: CMVerifySetting{
			Enabled:               true,
			WelcomeChannelId:      "1078237141139730452",
			ReactionRoleChannelId: "1078237160244793404",
		},
		QuestionDiscussionMoverSetting: QuestionDiscussionMoverSetting{
			Enabled:                   true,
			QuestionChannelId:         "1078336196373123194",
			AnsweredQuestionChannelId: "1078336218787479572",
			DiscussionAnswerSetting: []DiscussionAnswerSetting{
				{
					ChannelId:  "1078336179814027345",
					ReactionId: "✅",
				},
			},
			TriggerReactionId: "😂",
		},
		CMQuestionLimiterSetting: CMQuestionLimiterSetting{
			Enabled:                        true,
			ChannelId:                      "1078336196373123194",
			UnlimitedRoleIds:               map[string]bool{"1078336615958708384z": true},
			QuestionLimitDurationInMinutes: 1,
		},
		CMQuestionOneSetting: CMQuestionOneSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "1078237386615562253",
		},
		InvalidVettingResponseSetting: InvalidVettingResponseSetting{
			Enabled:                     true,
			Keyword:                     "stong",
			ChannelId:                   "1088370411470860288",
			VettingResponseKeywordFlags: []string{"inibajakan"},
		},
	}
}
