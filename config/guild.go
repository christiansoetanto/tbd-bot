package config

type GuildConfig struct {
	GuildId                                GuildId
	GuildName                              string
	Channel                                Channel
	Role                                   Role
	Moderator                              Moderator
	Reaction                               Reaction
	Wording                                Wording
	ReligionRoleMappingMap                 ReligionRoleMappingMap
	RegisteredFeature                      RegisteredFeature
	DetectVettingQuestioningKeywordSetting DetectVettingQuestioningKeywordSetting
}
type DetectVettingQuestioningKeywordSetting struct {
	Enabled     bool
	Keyword     string
	ChannelId   string
	Title       string
	Description string
}
type Channel struct {
	GeneralDiscussion            string
	ReactionRoles                string
	ServerInformation            string
	ReligiousQuestions           string
	ReligiousDiscussions1        string
	ReligiousDiscussions2        string
	AnsweredQuestions            string
	FAQ                          string
	Responses                    string
	VettingQuestioning           string
	RulesVetting                 string
	LiturgicalCalendarDiscussion string
	BotTesting                   string
}
type Moderator map[string]string
type Reaction struct {
	Upvote                                    string
	Dab                                       string
	ReligiousDiscussionOneWhiteCheckmark      string
	ReligiousDiscussionsTwoBallotBoxWithCheck string
}
type Role struct {
	Vetting            string
	VettingQuestioning string
	ApprovedUser       string
	LatinCatholic      string
	EasternCatholic    string
	OrthodoxChristian  string
	RCIACatechumen     string
	Protestant         string
	NonCatholic        string
	Atheist            string
	Moderator          string
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
type Wording struct {
	VerifyAckMessageFormat    string
	WelcomeMessageEmbedFormat string
	WelcomeMessageFormat      string
	MissedQuestionOneFormat   string
	WelcomeTitle              string
}

type GuildId string

var devGuild = GuildConfig{
	GuildId:   "813302330782253066",
	GuildName: "Local",
	Channel:   Channel{},
	Role: Role{
		Moderator: "1013781460953616404",
	},
	Moderator:              nil,
	Reaction:               Reaction{},
	Wording:                Wording{},
	ReligionRoleMappingMap: nil,
	RegisteredFeature:      nil,
	DetectVettingQuestioningKeywordSetting: DetectVettingQuestioningKeywordSetting{
		Enabled:     true,
		Keyword:     "keyword",
		ChannelId:   "1013780704330526834",
		Title:       "",
		Description: "You got the code right, <@%s>. Kindly wait for the mods to verify you.",
	},
}
