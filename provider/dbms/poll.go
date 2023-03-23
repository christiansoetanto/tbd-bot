package dbms

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
)

type Poll struct {
	Id       string   `json:"id,omitempty" firestore:"id"`
	Question string   `json:"question,omitempty" firestore:"question"`
	Options  []Option `json:"options,omitempty" firestore:"options"`
}

type Option struct {
	Value  string  `json:"value,omitempty" firestore:"value"`
	Weight int     `json:"weight,omitempty" firestore:"weight"`
	Voters []Voter `json:"voters,omitempty" firestore:"voters"`
}

type Voter struct {
	UserId string `json:"user_id,omitempty" firestore:"user_id"`
}

func (db *firestoreDb) UpsertPoll(ctx context.Context, data Poll) error {
	_, err := db.collection(polls).Doc(data.Id).Set(ctx, data)
	if err != nil {
		logv2.Error(ctx, err, data)
		return err
	}
	return nil
}
func (db *firestoreDb) GetPoll(ctx context.Context, id string) (Poll, error) {
	snap, err := db.collection(polls).Doc(id).Get(ctx)
	var poll Poll
	d := snap.Data()
	err = util.DecodeFirestore(d, &poll)
	if err != nil {
		logv2.Error(ctx, err, d)
		return Poll{}, err
	}
	return poll, nil
}
