package config

import "github.com/christiansoetanto/tbd-bot/domain"

func getTerraSanctaProdGuildConfig() GuildConfig {
	return GuildConfig{
		GuildId:   "800729899677646848",
		GuildName: "Terra Sancta",
		Channel: Channel{
			LiturgicalCalendarDiscussion: "945783888662388827",
			OfficeOfReadings:             "945783888662388827",
			RulesVetting:                 "806649992106213407",
			ChristianMemes:               "800735134189355009",
		},
		Role: Role{
			Vetting:            []string{"953168893101367336"},
			VettingQuestioning: "953168901443846154",
			Detained:           "929802827050655754",
			ApprovedUser:       "800733329451122688",
			Moderator:          "800852502357737494",
			ModeratorOrAbove:   []string{"800852502357737494", "929800261898211429", "942143857309655160"},
		},
		RegisteredFeature: map[string]bool{
			domain.FeatureKeyPing:                 true,
			domain.FeatureKeyTSVerify:             true,
			domain.FeatureKeyTSQuestionOne:        true,
			domain.FeatureKeyTSDetain:             true,
			domain.FeatureKeyTSOfficeOfReadings:   true,
			domain.FeatureKeyTSCalendar:           true,
			domain.FeatureKeyCalendarCron:         true,
			domain.FeatureKeyOfficeOfReadingsCron: true,
			domain.FeatureKeyFridayCron:           false,
		},
		TSVerifySetting: SDVerifySetting{
			Enabled: true,
			ReligionRoleMap: map[domain.ReligionRoleKey]string{
				domain.ReligionRoleKeyLatinCatholic:     "800733358660649000",
				domain.ReligionRoleKeyEasternCatholic:   "800779939191193640",
				domain.ReligionRoleKeyOrthodoxChristian: "800851942028083271",
				domain.ReligionRoleKeyRCIACatechumen:    "800852053433909288",
				domain.ReligionRoleKeyProtestant:        "800851975172259850",
				domain.ReligionRoleKeyNonCatholic:       "929804029389844541",
				domain.ReligionRoleKeyAtheist:           "804837240676679710",
			},
			WelcomeChannelId:           "800729899677646851",
			ReactionRoleChannelId:      "849042189988266075",
			ServerInformationChannelId: "993132727538819122",
		},
		QuestionDiscussionMoverSetting: QuestionDiscussionMoverSetting{
			Enabled:                   true,
			QuestionChannelId:         "945784003464679534",
			AnsweredQuestionChannelId: "993132917448523828",
			DiscussionAnswerSetting: []DiscussionAnswerSetting{
				{
					ChannelId:  "945783650849554472",
					ReactionId: "✅",
				},
				{
					ChannelId:  "945783731845754950",
					ReactionId: "☑️",
				},
			},
			TriggerReactionId: "949662090426204211",
		},
		TSQuestionOneSetting: SDQuestionOneSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "968540930791591996",
		},
		TSVettingQuestioningSetting: SDVettingQuestioningSetting{
			Enabled:                     true,
			VettingQuestioningChannelId: "968540930791591996",
		},
		InvalidVettingResponseSetting: InvalidVettingResponseSetting{
			Enabled:                     true,
			Keyword:                     "inri",
			ChannelId:                   "849052057582829568",
			VettingResponseKeywordFlags: []string{"whatcode", "andgiveusthecode"},
		},
	}

}
