package dbms

import (
	"github.com/christiansoetanto/tbd-bot/database"
	"sync"
)

type Obj struct {
	FirestoreDb FirestoreDb
}

var dbms *Obj
var once sync.Once

func GetDbmsClient(databaseObj *database.Obj) *Obj {
	once.Do(func() {
		connPayment, err := databaseObj.GetDb(database.FirestoreDb)
		if err != nil {
			panic(err)
		}
		dbms = &Obj{
			FirestoreDb: getFirestoreDbObj(connPayment),
		}
	})

	return dbms
}
