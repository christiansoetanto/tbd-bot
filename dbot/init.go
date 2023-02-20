package dbot

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/dbot/handler"
	"sync"
)

type Resource struct {
	Config  config.Config
	Session *discordgo.Session
	Handler handler.Handler
}

type usecase struct {
	*Resource
}

func (u *usecase) DoHelloWorld(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}

type Usecase interface {
	Init(ctx context.Context) error
	CloseDiscordgoConn() error
	DoHelloWorld(ctx context.Context)
}

var obj Usecase
var once sync.Once

func GetUsecaseObject(resource *Resource) Usecase {
	once.Do(func() {
		obj = &usecase{
			Resource: resource,
		}
	})
	return obj
}

func (u *usecase) Init(ctx context.Context) error {
	//handlers => open conn => cron jobs
	u.initHandlers(ctx)

	err := u.openDiscordgoConn()
	if err != nil {
		return err
	}
	//u.registerSlashCommand()
	//u.LoadAllCronJobs()

	return nil
}
func (u *usecase) openDiscordgoConn() error {
	u.Session.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMessageReactions | discordgo.IntentDirectMessages
	return u.Session.Open()
}
func (u *usecase) CloseDiscordgoConn() error {
	return u.Session.Close()
}

func (u *usecase) initHandlers(ctx context.Context) {
	for _, h := range u.Handler.GetHandlers(ctx) {
		u.Session.AddHandler(h)
	}
}
