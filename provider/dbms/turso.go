package dbms

import (
	"context"
	"database/sql"
)

type Turso interface {
	InsertRoles(ctx context.Context, userId string, guildId string, roles []string) error
	GetRoles(ctx context.Context, userId string, guildId string) ([]string, error)
}
type turso struct {
	c *sql.DB
}

func (t *turso) GetRoles(ctx context.Context, userId string, guildId string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (t *turso) InsertRoles(ctx context.Context, userId string, guildId string, roles []string) error {
	//TODO implement me
	panic("implement me")
}

func getTursoObj(db *sql.DB) Turso {
	return &turso{
		c: db,
	}
}
