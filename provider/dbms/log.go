package dbms

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
)

type LogActivity struct {
	UserId   string `json:"user_id,omitempty" firestore:"user_id"`
	Username string `json:"username,omitempty" firestore:"username"`
	Activity string `json:"activity,omitempty" firestore:"activity"`
}

func (db *firestoreDb) InsertLogActivity(ctx context.Context, logActivity LogActivity) error {
	ctx = logv2.InitFuncContext(ctx)
	logv2.Debug(ctx, logv2.Request, logActivity)
	_, _, err := db.getCollection(logs).Add(ctx, logActivity)
	if err != nil {
		logv2.Error(ctx, err)
		return err
	}
	return err
}
