package dbot

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/dbot/handler"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/robfig/cron/v3"
	"os"
	"sync"
	"time"
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
	util.InitMetrics()
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

// sessionStater exposes discordgo's heartbeat bookkeeping to util without
// util having to import discordgo. LastHeartbeatAck is a plain field guarded
// by the session's own embedded mutex, so it has to be read under that lock.
type sessionStater struct {
	s *discordgo.Session
}

func (a sessionStater) LastHeartbeatAck() time.Time {
	a.s.RLock()
	defer a.s.RUnlock()
	return a.s.LastHeartbeatAck
}

func (a sessionStater) HeartbeatLatency() time.Duration {
	return a.s.HeartbeatLatency()
}

func (u *usecase) openDiscordgoConn() error {
	u.Session.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMessageReactions | discordgo.IntentDirectMessages
	u.registerGatewayObservers()
	return u.Session.Open()
}

// registerGatewayObservers records gateway lifecycle transitions and points
// util at this session's heartbeat clock. The counters describe what happened;
// the heartbeat clock is what decides whether the bot is alive, because a
// connection can wedge without ever emitting Disconnect.
func (u *usecase) registerGatewayObservers() {
	util.SetGatewayStater(sessionStater{s: u.Session})

	u.Session.AddHandler(func(s *discordgo.Session, _ *discordgo.Connect) {
		util.SetDiscordConnected(true)
		util.IncGatewayEvent("connect")
	})
	u.Session.AddHandler(func(s *discordgo.Session, _ *discordgo.Disconnect) {
		util.SetDiscordConnected(false)
		util.IncGatewayEvent("disconnect")
	})
	u.Session.AddHandler(func(s *discordgo.Session, _ *discordgo.Resumed) {
		util.SetDiscordConnected(true)
		util.IncGatewayEvent("resumed")
	})
	u.Session.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		util.SetDiscordConnected(true)
		util.IncGatewayEvent("ready")
	})
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

// heartbeatCronJob reports gateway liveness to an external dead-man's switch.
// Prometheus and Grafana run on the same Mac Mini as the bot, so nothing here
// can report that machine losing power, network or Colima — the alerter dies
// with the thing it watches, and silence reads as health.
func (u *usecase) heartbeatCronJob(ctx context.Context) func() {
	heartbeat := util.NewHeartbeat(os.Getenv("HEALTHCHECKS_PING_URL"))
	if !heartbeat.Enabled() {
		logv2.Error(ctx, errors.New("HEALTHCHECKS_PING_URL is not set"),
			"external heartbeat disabled: a host-down outage will not alert")
	}
	return func() {
		util.IncCronExecutions("heartbeat")
		heartbeat.Ping(ctx)
	}
}

func (u *usecase) loadAllCronJobs(ctx context.Context) {
	const DailyCron = "@daily"
	const FridayCron = "0 0 * * 5"
	const Every5SecondCron = "@every 5s"
	const HeartbeatCron = "@every 1m"
	success := 0
	c := cron.New()
	_, err := c.AddFunc(HeartbeatCron, u.heartbeatCronJob(ctx))
	if err != nil {
		logv2.Error(ctx, err, "external heartbeat cron job failed to load")
	} else {
		success++
	}
	_, err = c.AddFunc(DailyCron, u.liturgicalCalendarCronJob(ctx))
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
