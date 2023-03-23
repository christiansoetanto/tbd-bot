package dbms

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"google.golang.org/api/iterator"
	"time"
)

type Question struct {
	UserId  string    `json:"user_id,omitempty" firestore:"user_id"`
	GuildId string    `json:"guild_id,omitempty" firestore:"guild_id"`
	Time    time.Time `json:"time" firestore:"time"`
	DocId   string    `json:"doc_id,omitempty" firestore:"doc_id"`
}

func (db *firestoreDb) UpdateQuestionTime(ctx context.Context, docId string) error {
	_, err := db.collection(questions).Doc(docId).Update(ctx, []firestore.Update{
		{
			Path:  "time",
			Value: time.Now(),
		},
	})
	if err != nil {
		logv2.Error(ctx, err, docId)
		return err
	}
	return nil
}

func (db *firestoreDb) InsertLatestQuestion(ctx context.Context, userId string, guildId string) error {
	data := Question{
		UserId:  userId,
		GuildId: guildId,
		Time:    time.Now(),
	}
	_, _, err := db.collection(questions).Add(ctx, data)
	if err != nil {
		logv2.Error(ctx, err, data)
		return err
	}
	return nil
}

func (db *firestoreDb) GetLatestQuestion(ctx context.Context, userId string, guildId string) (Question, error) {
	iter := db.collection(questions).Where("user_id", "==", userId).Where("guild_id", "==", guildId).OrderBy("time", firestore.Desc).Limit(1).Documents(ctx)
	for {
		var q Question
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			logv2.Error(ctx, err)
			return q, err
		}
		data := doc.Data()
		err = util.DecodeFirestore(data, &q)
		if err != nil {
			logv2.Error(ctx, err, data)
			return q, err
		}
		q.DocId = doc.Ref.ID
		logv2.Debug(ctx, logv2.Info, q)
		return q, nil
	}
	return Question{}, nil
}
