package provider

import (
	"context"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/database"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/provider/dbms"
	"sync"
)

type provider struct {
	Dbms      *dbms.Obj
	AppConfig config.AppConfig
}

type Resource struct {
	AppConfig config.AppConfig
	Database  *database.Obj
}

type Provider interface {
	HelloWorld(ctx context.Context) error
	UpsertLatestQuestion(ctx context.Context, q domain.Question) error
	GetLatestQuestion(ctx context.Context, userId string, guildId string) (domain.Question, error)
	GetPoll(ctx context.Context, pollId string) (domain.Poll, error)
	UpsertPoll(ctx context.Context, poll domain.Poll) error
}

var obj Provider
var once sync.Once

// GetProvider get provider client
func GetProvider(resource *Resource) Provider {
	once.Do(func() {
		obj = &provider{
			Dbms:      dbms.GetDbmsClient(resource.Database),
			AppConfig: resource.AppConfig,
		}
	})
	return obj
}
