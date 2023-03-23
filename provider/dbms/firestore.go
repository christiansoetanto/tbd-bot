package dbms

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"google.golang.org/api/iterator"
	"log"
	"os"
)

type FirestoreDb interface {
	HelloWorld(ctx context.Context) error
	InsertLogActivity(ctx context.Context, logActivity LogActivity) error
	GetLogActivity(ctx context.Context) error
	InsertLatestQuestion(ctx context.Context, userId string, guildId string) error
	UpdateQuestionTime(ctx context.Context, docId string) error
	GetLatestQuestion(ctx context.Context, userId string, guildId string) (Question, error)
	UpsertPoll(ctx context.Context, data Poll) error
	GetPoll(ctx context.Context, id string) (Poll, error)
}
type firestoreDb struct {
	c *firestore.Client
}

func getFirestoreDbObj(c *firestore.Client) FirestoreDb {
	return &firestoreDb{
		c: c,
	}
}

const (
	users     collectionType = "users"
	logs                     = "logs"
	questions                = "questions"
	polls                    = "polls"
)

type collectionType string

func (db *firestoreDb) collection(collection collectionType) *firestore.CollectionRef {
	env := os.Getenv("TBDENV")
	col := string(collection)
	if env == "staging" {
		col = col + "_staging"
	}
	return db.c.Collection(col)
}
func (db *firestoreDb) HelloWorld(ctx context.Context) error {
	iter := db.collection(users).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			log.Fatal(err)
			return err
		}
		if err == iterator.Done {
			break
		}
		logv2.Debug(ctx, logv2.Info, "User: %v", doc.Data())
	}
	return nil
}
