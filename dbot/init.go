package dbot

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/dbot/handler"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/robfig/cron/v3"
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
	u.registerSlashCommand(ctx)
	u.loadAllCronJobs(ctx)

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
func (u *usecase) registerSlashCommand(ctx context.Context) {
	commands, _ := u.Handler.GetCommandHandlers(ctx)
	for guildId, guild := range u.Config.GuildConfig {
		var guildCommands []*discordgo.ApplicationCommand
		for _, command := range commands {
			if guild.RegisteredFeature[command.Name] {
				guildCommands = append(guildCommands, command)
			}
		}
		_, err := u.Session.ApplicationCommandBulkOverwrite(u.Session.State.User.ID, string(guildId), guildCommands)
		if err != nil {
			logv2.Error(ctx, err, fmt.Sprintf("Cannot create command in guild %s: %v", string(guildId), guildCommands))
		}
	}
}

func (u *usecase) loadAllCronJobs(ctx context.Context) {
	const DailyCron = "@daily"
	const FridayCron = "0 0 * * 5"
	const Every5SecondCron = "@every 5s"
	success := 0
	c := cron.New()
	_, err := c.AddFunc(DailyCron, u.liturgicalCalendarCronJob(ctx))
	if err != nil {
		logv2.Error(ctx, err, "liturgical calendar cron job failed to load")
	} else {
		success++
	}
	_, err = c.AddFunc(DailyCron, u.officeOfReadingsCronJob(ctx))
	if err != nil {
		logv2.Error(ctx, err, "office of readings cron job failed to load")
	} else {
		success++
	}

	_, err = c.AddFunc(DailyCron, u.officeOfReadingsCronJob2(ctx))
	if err != nil {
		logv2.Error(ctx, err, "office of readings cron job 2 failed to load")
	} else {
		success++
	}
	_, err = c.AddFunc(FridayCron, u.fridayCronJob(ctx))
	if err != nil {
		logv2.Error(ctx, err, "friday cron job failed to load")
	} else {
		success++
	}

	if success != 0 {
		c.Start()
	} else {
		logv2.Error(ctx, errors.New("no cron job loaded"))
		return
	}
}
