package domain

import "time"

type FeatureKey string

const (
	FeatureKeyPing                 string = "ping"
	FeatureKeySDVerify             string = "sdverify"
	FeatureKeySDQuestionOne        string = "sdquestionone"
	FeatureKeySDVettingQuestioning string = "sdvettingquestioning"
	FeatureKeySDDetain             string = "sddetain"
	FeatureKeyCMVerify             string = "cmverify"
	FeatureKeyCMQuestionOne        string = "cmquestionone"
	FeatureKeyCMPoll               string = "cmpoll"
	FeatureKeyCalendarCron         string = "calendar_cron"
	FeatureKeyOfficeOfReadingsCron string = "office_of_readings_cron"
)

const (
	ComponentKeyCMVote       string = "vote"
	ComponentKeyCMShowVoters string = "show_voters"
)

type ReligionRoleKey string

const (
	ReligionRoleKeyLatinCatholic     ReligionRoleKey = "latin_catholic"
	ReligionRoleKeyEasternCatholic   ReligionRoleKey = "eastern_catholic"
	ReligionRoleKeyRCIACatechumen    ReligionRoleKey = "rcia_catechumen"
	ReligionRoleKeyOrthodoxChristian ReligionRoleKey = "orthodox_christian"
	ReligionRoleKeyProtestant        ReligionRoleKey = "protestant"
	ReligionRoleKeyNonCatholic       ReligionRoleKey = "non_catholic"
	ReligionRoleKeyAtheist           ReligionRoleKey = "atheist"
)

type Question struct {
	UserId  string    `json:"user_id,omitempty"`
	GuildId string    `json:"guild_id,omitempty"`
	Time    time.Time `json:"time"`
	DocId   string    `json:"doc_id,omitempty"`
}

type Poll struct {
	Id       string   `json:"id,omitempty" `
	Question string   `json:"question,omitempty" `
	Options  []Option `json:"options,omitempty"`
}

type Option struct {
	Value  string  `json:"value,omitempty"`
	Weight int     `json:"weight,omitempty"`
	Voters []Voter `json:"voters,omitempty"`
}

type Voter struct {
	UserId string `json:"user_id,omitempty"`
}
