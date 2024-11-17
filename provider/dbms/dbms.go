package dbms

import (
	"github.com/christiansoetanto/tbd-bot/database"
	"sync"
)

type Obj struct {
	FirestoreDb FirestoreDb
	TursoDb     Turso
}

var dbms *Obj
var once sync.Once

func GetDbmsClient(databaseObj *database.Obj) *Obj {
	once.Do(func() {
		connFirestore, err := databaseObj.GetDb(database.FirestoreDb)
		if err != nil {
			panic(err)
		}

		connTurso, err := databaseObj.GetTursoDb(database.TursoDb)
		if err != nil {
			panic(err)
		}
		dbms = &Obj{
			FirestoreDb: getFirestoreDbObj(connFirestore),
			TursoDb:     getTursoObj(connTurso),
		}
	})

	return dbms
}
