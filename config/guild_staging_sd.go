package config

import "github.com/christiansoetanto/tbd-bot/domain"

func getServusDeiStagingGuildConfig() GuildConfig {
	return GuildConfig{
		GuildId:   "813302330782253066",
		GuildName: "Dev - Servus Dei",
		Channel: Channel{
			LiturgicalCalendarDiscussion: "1013780907192221757",
			OfficeOfReadings:             "1013780907192221757",
			RulesVetting:                 "1013780880591954002",
			ChristianMemes:               "1129281996053565482",
		},
		Role: Role{
			Vetting:            []string{"974632148952809482"},
			VettingQuestioning: "974632188823863296",
			Detained:           "1129260016663277579",
			ApprovedUser:       "974632216304943155",
			Moderator:          "1013781460953616404",
			ModeratorOrAbove:   []string{"1013781460953616404"},
		},
		Reaction: Reaction{},
		RegisteredFeature: map[string]bool{
			domain.FeatureKeyPing:                 true,
			domain.FeatureKeySDVerify:             true,
			domain.FeatureKeySDQuestionOne:        true,
			domain.FeatureKeySDVettingQuestioning: true,
			domain.FeatureKeySDDetain:             true,
			domain.FeatureKeyCalendarCron:         true,
			domain.FeatureKeyOfficeOfReadingsCron: true,
			domain.FeatureKeyFridayCron:           true,
		},
		DetectVettingQuestioningKeywordSetting: DetectVettingQuestioningKeywordSetting{
			Enabled:     true,
			Keyword:     "keyword",
			ChannelId:   "1013780704330526834",
			Title:       "",
			Description: "You got the code right, <@%s>. Kindly wait for the mods to verify you.",
		},
		SDVerifySetting: SDVerifySetting{
			Enabled: true,
			ReligionRoleMap: map[domain.ReligionRoleKey]string{
				domain.ReligionRoleKeyLatinCatholic:     "974630535395680337",
				domain.ReligionRoleKeyEasternCatholic:   "974667212587671613",
				domain.ReligionRoleKeyOrthodoxChristian: "974667248826449950",
				domain.ReligionRoleKeyRCIACatechumen:    "974667251498225704",
				domain.ReligionRoleKeyProtestant:        "974667253045919784",
				domain.ReligionRoleKeyNonCatholic:       "974667254627201084",
				domain.ReligionRoleKeyAtheist:           "974667257122795570",
			},
			WelcomeChannelId:           "1013780724345745508",
			ReactionRoleChannelId:      "1013780802619854848",
			ServerInformationChannelId: "1013780836203638836",
		},
		QuestionDiscussionMoverSetting: QuestionDiscussionMoverSetting{
			Enabled:                   true,
			QuestionChannelId:         "1013780754834145333",
			AnsweredQuestionChannelId: "1013780765307310091",
			DiscussionAnswerSetting: []DiscussionAnswerSetting{
				{
					ChannelId:  "1013780733510287472",
					ReactionId: "✅",
				},
				{
					ChannelId:  "1013780741542379620",
					ReactionId: "☑️",
				},
			},
			TriggerReactionId: "1013782200052887683",
		},
		SDQuestionOneSetting: SDQuestionOneSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "1013780704330526834",
		},
		SDVettingQuestioningSetting: SDVettingQuestioningSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "914987511481249792",
		},
		InvalidVettingResponseSetting: InvalidVettingResponseSetting{
			Enabled:                     true,
			Keyword:                     "keyword",
			ChannelId:                   "1013780662798528592",
			VettingResponseKeywordFlags: []string{"inibajakan"},
		},
	}
}
