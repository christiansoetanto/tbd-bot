package config

import (
	"github.com/christiansoetanto/tbd-bot/domain"
)

type GuildConfig struct {
	GuildId           GuildId
	GuildName         string
	Channel           Channel
	Role              Role
	Moderator         Moderator
	Reaction          Reaction
	RegisteredFeature RegisteredFeature

	//start common

	DetectVettingQuestioningKeywordSetting DetectVettingQuestioningKeywordSetting
	QuestionDiscussionMoverSetting         QuestionDiscussionMoverSetting
	InvalidVettingResponseSetting          InvalidVettingResponseSetting

	//end common

	// ==========

	//start SD

	SDVerifySetting      SDVerifySetting
	SDQuestionOneSetting SDQuestionOneSetting

	//end SD

	// =============================

	//start CM

	CMQuestionOneSetting     CMQuestionOneSetting
	CMVerifySetting          CMVerifySetting
	CMQuestionLimiterSetting CMQuestionLimiterSetting

	//end CM
}
type CMQuestionOneSetting struct {
	Enabled                     bool   `json:"enabled,omitempty"`
	VettingQuestioningChannelId string `json:"vetting_questioning_channel_id,omitempty"`
}
type SDQuestionOneSetting struct {
	Enabled                     bool   `json:"enabled,omitempty"`
	VettingQuestioningChannelId string `json:"vetting_questioning_channel_id,omitempty"`
}
type CMQuestionLimiterSetting struct {
	Enabled                        bool            `json:"enabled,omitempty"`
	ChannelId                      string          `json:"channel_id,omitempty"`
	UnlimitedRoleIds               map[string]bool `json:"unlimited_role_ids,omitempty"`
	QuestionLimitDurationInMinutes int             `json:"question_limit_duration_in_minutes,omitempty"`
}
type QuestionDiscussionMoverSetting struct {
	Enabled                   bool                      `json:"enabled,omitempty"`
	QuestionChannelId         string                    `json:"question_channel_id,omitempty"`
	AnsweredQuestionChannelId string                    `json:"answered_question_channel_id,omitempty"`
	DiscussionAnswerSetting   []DiscussionAnswerSetting `json:"discussion_answer_setting,omitempty"`
	TriggerReactionId         string                    `json:"reaction_id,omitempty"`
}
type DiscussionAnswerSetting struct {
	ChannelId  string `json:"channel_id,omitempty"`
	ReactionId string `json:"reaction_id,omitempty"`
}
type SDVerifySetting struct {
	Enabled                    bool                              `json:"enabled,omitempty"`
	ReligionRoleMap            map[domain.ReligionRoleKey]string `json:"religion_role_map,omitempty"`
	WelcomeChannelId           string                            `json:"welcome_channel_id,omitempty"`
	ReactionRoleChannelId      string                            `json:"reaction_role_channel_id,omitempty"`
	ServerInformationChannelId string                            `json:"server_information_channel_id,omitempty"`
}
type CMVerifySetting struct {
	Enabled               bool   `json:"enabled,omitempty"`
	WelcomeChannelId      string `json:"welcome_channel_id,omitempty"`
	ReactionRoleChannelId string `json:"reaction_role_channel_id,omitempty"`
}
type SDVerifyChoice struct {
	Name   string `json:"name,omitempty"`
	Key    string `json:"key,omitempty"`
	RoleId string `json:"role_id,omitempty"`
}
type DetectVettingQuestioningKeywordSetting struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	ChannelId   string `json:"channel_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
type InvalidVettingResponseSetting struct {
	Enabled                     bool     `json:"enabled,omitempty"`
	Keyword                     string   `json:"keyword,omitempty"`
	ChannelId                   string   `json:"channel_id,omitempty"`
	VettingResponseKeywordFlags []string `json:"vetting_response_keyword_flags,omitempty"`
}

type Channel struct {
	GeneralDiscussion            string `json:"general_discussion,omitempty"`
	ReactionRoles                string `json:"reaction_roles,omitempty"`
	ServerInformation            string `json:"server_information,omitempty"`
	ReligiousQuestions           string `json:"religious_questions,omitempty"`
	ReligiousDiscussions1        string `json:"religious_discussions_1,omitempty"`
	ReligiousDiscussions2        string `json:"religious_discussions_2,omitempty"`
	AnsweredQuestions            string `json:"answered_questions,omitempty"`
	FAQ                          string `json:"faq,omitempty"`
	Responses                    string `json:"responses,omitempty"`
	RulesVetting                 string `json:"rules_vetting,omitempty"`
	LiturgicalCalendarDiscussion string `json:"liturgical_calendar_discussion,omitempty"`
	BotTesting                   string `json:"bot_testing,omitempty"`
}
type Moderator map[string]string
type Reaction struct {
	Gold string `json:"gold,omitempty"`
}
type Role struct {
	Vetting            []string `json:"vetting,omitempty"`
	VettingQuestioning string   `json:"vetting_questioning,omitempty"`
	ApprovedUser       string   `json:"approved_user,omitempty"`
	LatinCatholic      string   `json:"latin_catholic,omitempty"`
	EasternCatholic    string   `json:"eastern_catholic,omitempty"`
	OrthodoxChristian  string   `json:"orthodox_christian,omitempty"`
	RCIACatechumen     string   `json:"rcia_catechumen,omitempty"`
	Protestant         string   `json:"protestant,omitempty"`
	NonCatholic        string   `json:"non_catholic,omitempty"`
	Atheist            string   `json:"atheist,omitempty"`
	Moderator          string   `json:"moderator,omitempty"`
	ModeratorOrAbove   []string `json:"moderator_or_above,omitempty"`
	CMGoldMember       []string `json:"cm_gold_member,omitempty"`
}

const (
	LatinCatholic     ReligionRoleType = "Latin Catholic"
	EasternCatholic   ReligionRoleType = "Eastern Catholic"
	OrthodoxChristian ReligionRoleType = "Orthodox Christian"
	RCIACatechumen    ReligionRoleType = "RCIA / Catechumen"
	Protestant        ReligionRoleType = "Protestant"
	NonCatholic       ReligionRoleType = "Non Catholic"
	Atheist           ReligionRoleType = "Atheist"
)

type ReligionRoleType string
type ReligionRoleId string
type ReligionRoleMappingMap map[ReligionRoleType]ReligionRoleId

type GuildId string

var devServusDeiGuild = GuildConfig{
	GuildId:   "813302330782253066",
	GuildName: "Dev - Servus Dei",
	Channel: Channel{
		LiturgicalCalendarDiscussion: "1013780907192221757",
		RulesVetting:                 "1013780880591954002",
	},
	Role: Role{
		Vetting:            []string{"974632148952809482"},
		VettingQuestioning: "974632188823863296",
		ApprovedUser:       "974632216304943155",
		Moderator:          "1013781460953616404",
		ModeratorOrAbove:   []string{"1013781460953616404"},
	},
	Moderator: nil,
	Reaction:  Reaction{},
	RegisteredFeature: map[string]bool{
		domain.FeatureKeyPing:          true,
		domain.FeatureKeySDVerify:      true,
		domain.FeatureKeySDQuestionOne: true,
		domain.FeatureKeyCalendarCron:  true,
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
	InvalidVettingResponseSetting: InvalidVettingResponseSetting{
		Enabled:                     true,
		Keyword:                     "keyword",
		ChannelId:                   "1013780662798528592",
		VettingResponseKeywordFlags: []string{"inibajakan"},
	},
}

var devCapitalMindsetGuild = GuildConfig{
	GuildId:   "1077804660997500980",
	GuildName: "dev - capital mindset",
	Channel: Channel{
		GeneralDiscussion:            "",
		ReactionRoles:                "",
		ServerInformation:            "",
		ReligiousQuestions:           "",
		ReligiousDiscussions1:        "",
		ReligiousDiscussions2:        "",
		AnsweredQuestions:            "",
		FAQ:                          "",
		Responses:                    "",
		RulesVetting:                 "1088360872558211164",
		LiturgicalCalendarDiscussion: "",
		BotTesting:                   "",
	},
	Role: Role{
		Vetting:            []string{"1078236690004594759", "1078236722405593118"},
		VettingQuestioning: "1078241164681019462",
		ApprovedUser:       "1078236749861494807",
		LatinCatholic:      "",
		EasternCatholic:    "",
		OrthodoxChristian:  "",
		RCIACatechumen:     "",
		Protestant:         "",
		NonCatholic:        "",
		Atheist:            "",
		Moderator:          "1078336581703843840",
		ModeratorOrAbove:   []string{"1078336581703843840", "1078336615958708384"},
		CMGoldMember:       []string{"1078236771751563305", "1078236793402576907"},
	},
	Moderator: nil,
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
	SDVerifySetting: SDVerifySetting{
		Enabled:                    false,
		ReligionRoleMap:            nil,
		WelcomeChannelId:           "",
		ReactionRoleChannelId:      "",
		ServerInformationChannelId: "",
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
