package config

import "github.com/christiansoetanto/tbd-bot/domain"

func getServusDeiProdGuildConfig() GuildConfig {
	return GuildConfig{
		GuildId:   "751139261515825162",
		GuildName: "Servus Dei",
		Channel: Channel{
			LiturgicalCalendarDiscussion: "915621423270211594",
			OfficeOfReadings:             "1103015883955261532",
			RulesVetting:                 "775654889934159893",
		},
		Role: Role{
			Vetting:            []string{"751145124834312342"},
			VettingQuestioning: "914986915030241301",
			ApprovedUser:       "751144797938384979",
			Moderator:          "751144316843327599",
			ModeratorOrAbove:   []string{"751144316843327599", "802559328220872704", "751144128133201920"},
		},
		RegisteredFeature: map[string]bool{
			domain.FeatureKeyPing:                 true,
			domain.FeatureKeySDVerify:             true,
			domain.FeatureKeySDQuestionOne:        true,
			domain.FeatureKeySDVettingQuestioning: true,
			domain.FeatureKeyCalendarCron:         true,
			domain.FeatureKeyOfficeOfReadingsCron: true,
		},
		DetectVettingQuestioningKeywordSetting: DetectVettingQuestioningKeywordSetting{
			Enabled:     true,
			Keyword:     "inri",
			ChannelId:   "914987511481249792",
			Description: "You got the code right, <@%s>. Kindly wait for the mods to verify you.",
		},
		SDVerifySetting: SDVerifySetting{
			Enabled: true,
			ReligionRoleMap: map[domain.ReligionRoleKey]string{
				domain.ReligionRoleKeyLatinCatholic:     "751145824532168775",
				domain.ReligionRoleKeyEasternCatholic:   "751148911267414067",
				domain.ReligionRoleKeyOrthodoxChristian: "751148354716565656",
				domain.ReligionRoleKeyRCIACatechumen:    "751196794771472395",
				domain.ReligionRoleKeyProtestant:        "751145951137103872",
				domain.ReligionRoleKeyNonCatholic:       "751146099351224382",
				domain.ReligionRoleKeyAtheist:           "751148904938209351",
			},
			WelcomeChannelId:           "751174152588623912",
			ReactionRoleChannelId:      "767452241321000970",
			ServerInformationChannelId: "973586981789499452",
		},
		QuestionDiscussionMoverSetting: QuestionDiscussionMoverSetting{
			Enabled:                   true,
			QuestionChannelId:         "751174501307383908",
			AnsweredQuestionChannelId: "821657995129126942",
			DiscussionAnswerSetting: []DiscussionAnswerSetting{
				{
					ChannelId:  "751174442217898065",
					ReactionId: "✅",
				},
				{
					ChannelId:  "771836244879605811",
					ReactionId: "☑️",
				},
			},
			TriggerReactionId: "762045856592822342",
		},
		SDQuestionOneSetting: SDQuestionOneSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "914987511481249792",
		},
		SDVettingQuestioningSetting: SDVettingQuestioningSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "914987511481249792",
		},
		InvalidVettingResponseSetting: InvalidVettingResponseSetting{
			Enabled:                     true,
			Keyword:                     "inri",
			ChannelId:                   "751151421231202363",
			VettingResponseKeywordFlags: []string{"whatcode", "andgiveusthecode"},
		},
	}

}
