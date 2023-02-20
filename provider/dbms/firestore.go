package dbms

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"google.golang.org/api/iterator"
	"log"
)

type FirestoreDb interface {
	HelloWorld(ctx context.Context) error
	InsertLogActivity(ctx context.Context, logActivity LogActivity) error
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
	users collectionType = "users"
	logs                 = "logs"
)

type collectionType string

func (db *firestoreDb) getCollection(collection collectionType) *firestore.CollectionRef {
	return db.c.Collection(string(collection))
}
func (db *firestoreDb) HelloWorld(ctx context.Context) error {
	iter := db.getCollection(users).Documents(ctx)
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
