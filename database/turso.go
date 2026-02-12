package database

import (
	"database/sql"
	"errors"
)

const (
	TursoDb DBType = "TursoDb"
)

func (r *Obj) GetTursoDb(dType DBType) (conn *sql.DB, err error) {
	obj.mtx.RLock()
	defer obj.mtx.RUnlock()
	if dbConn, ok := obj.tursoDbs[dType]; ok {
		return dbConn, nil
	}
	return nil, errors.New("no db found")
}
