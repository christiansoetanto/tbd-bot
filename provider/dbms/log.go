package dbms

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"google.golang.org/api/iterator"
)

type LogActivity struct {
	UserId    string                 `json:"user_id,omitempty" firestore:"user_id"`
	Username  string                 `json:"username,omitempty" firestore:"username"`
	Activity  string                 `json:"activity,omitempty" firestore:"activity"`
	RandomMap map[string]interface{} `json:"random_map,omitempty" firestore:"random_map"`
}

func (db *firestoreDb) InsertLogActivity(ctx context.Context, logActivity LogActivity) error {
	ctx = logv2.InitFuncContext(ctx)
	logv2.Debug(ctx, logv2.Request, logActivity)
	_, _, err := db.collection(logs).Add(ctx, logActivity)
	if err != nil {
		logv2.Error(ctx, err)
		return err
	}
	return err
}

func (db *firestoreDb) GetLogActivity(ctx context.Context) error {
	ctx = logv2.InitFuncContext(ctx)
	iter := db.collection(logs).Documents(ctx)
	var qs []LogActivity
	for {
		var q LogActivity
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			logv2.Error(ctx, err)
			return err
		}
		data := doc.Data()
		err = util.DecodeFirestore(data, &q)
		if err != nil {
			logv2.Error(ctx, err, data)
			return err
		}
		qs = append(qs, q)
	}
	logv2.Debug(ctx, logv2.Info, qs)
	return nil
}
