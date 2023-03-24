package config

import (
	"github.com/christiansoetanto/tbd-bot/domain"
	"os"
)

type GuildConfig struct {
	GuildId           GuildId
	GuildName         string
	Channel           Channel
	Role              Role
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
	RulesVetting                 string `json:"rules_vetting,omitempty"`
	LiturgicalCalendarDiscussion string `json:"liturgical_calendar_discussion,omitempty"`
}
type Reaction struct {
	Gold string `json:"gold,omitempty"`
}
type Role struct {
	Vetting            []string `json:"vetting,omitempty"`
	VettingQuestioning string   `json:"vetting_questioning,omitempty"`
	ApprovedUser       string   `json:"approved_user,omitempty"`
	Moderator          string   `json:"moderator,omitempty"`
	ModeratorOrAbove   []string `json:"moderator_or_above,omitempty"`
	CMGoldMember       []string `json:"cm_gold_member,omitempty"`
}

type GuildId string

func buildGuildConfig() map[GuildId]GuildConfig {
	cfg := make(map[GuildId]GuildConfig)
	env := os.Getenv("TBDENV")
	var sd GuildConfig
	var cm GuildConfig
	if env == "staging" {
		sd = getServusDeiStagingGuildConfig()
		cm = getCapitalMindsetStagingGuildConfig()
	} else {
		sd = getServusDeiProdGuildConfig()
		cm = getCapitalMindsetProdGuildConfig()
	}
	cfg[sd.GuildId] = sd
	cfg[cm.GuildId] = cm
	return cfg
}
