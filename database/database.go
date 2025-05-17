package database

import (
	"cloud.google.com/go/firestore"
	"context"
	"database/sql"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"google.golang.org/api/option"
	"log"
	"sync"
)

type DBType string

type Obj struct {
	connectedDbs map[DBType]*firestore.Client
	tursoDbs     map[DBType]*sql.DB
	mtx          sync.RWMutex
}

var obj *Obj
var once sync.Once

const (
	FirestoreDb DBType = "FirestoreDb"
)

func New(ctx context.Context, cfg config.AppConfig) {
	once.Do(func() {
		err := initDb(ctx, cfg)
		if err != nil {
			log.Fatal("Failed to init database", err)
		}
	})
}
func initDb(ctx context.Context, cfg config.AppConfig) error {
	obj = &Obj{connectedDbs: make(map[DBType]*firestore.Client)}

	if cfg.FirebaseServiceAccountJson != "" {
		sa := option.WithCredentialsJSON([]byte(cfg.FirebaseServiceAccountJson))
		app, err := firebase.NewApp(ctx, nil, sa)
		if err != nil {
			log.Fatalln(err)
		}

		client, err := app.Firestore(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		logv2.Debug(ctx, logv2.Info, "Firestore client created")
		obj.connectedDbs[FirestoreDb] = client
	}
	if cfg.TursoAuth != "" {
		url := fmt.Sprintf("libsql://%s.turso.io?authToken=%s", cfg.TursoName, cfg.TursoAuth)
		db, err := sql.Open("libsql", url)
		if err != nil {
			log.Fatalln(err)
		}
		obj.tursoDbs[TursoDb] = db
	}
	return nil
}

func Close(ctx context.Context) {
	for _, db := range obj.connectedDbs {
		err := db.Close()
		if err != nil {
			log.Fatal("Failed to close database", err)
		}
	}
	for _, db := range obj.tursoDbs {
		err := db.Close()
		if err != nil {
			log.Fatal("Failed to close database", err)
		}
	}
}
func GetDBObject(ctx context.Context, cfg config.AppConfig) *Obj {
	if obj == nil {
		New(ctx, cfg)
	}
	return obj
}
func (r *Obj) GetDb(dType DBType) (conn *firestore.Client, err error) {
	obj.mtx.RLock()
	defer obj.mtx.RUnlock()
	if dbConn, ok := obj.connectedDbs[dType]; ok {
		return dbConn, nil
	}
	return nil, errors.New("no db found")
}
