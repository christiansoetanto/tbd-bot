package logv2

import (
	"context"
	"fmt"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/rs/zerolog"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger zerolog.Logger
	once   sync.Once
)

func Init(cfg config.AppConfig) {
	once.Do(func() {
		writer := os.Stderr
		if cfg.DevMode {
			path := "/var/log/tbd-bot/tbdbot.log"
			err := os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil && err != os.ErrExist {
				log.Fatal(err)
			}
			writer, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		logger = zerolog.New(writer).
			With().Stack().Timestamp().CallerWithSkipFrameCount(3).Str("app", "tbd-bot").Logger()

	})
}

type ContextMetadata struct {
	Func      string `json:"func,omitempty"`
	Flow      string `json:"flow,omitempty"`
	TimeSpent int64  `json:"time_spent,omitempty"`
}

type DebugType string

const (
	Info     DebugType = "Info"
	Warning  DebugType = "Warning"
	Response DebugType = "Response"
	Request  DebugType = "Request"
)
const Finish string = "Finished process without error"

func withContext(ctx context.Context) ContextMetadata {
	return ContextMetadata{
		Func:      getCtxFunc(ctx),
		Flow:      getCtxFlow(ctx),
		TimeSpent: getCtxTimeSpent(ctx),
	}
}

func Debug(ctx context.Context, debugType DebugType, data interface{}, additionalInfo ...interface{}) {
	var msg string
	if len(additionalInfo) > 0 {
		msg = fmt.Sprintf("%s : %s", additionalInfo, util.InterfaceToJSON(data))
	} else {
		msg = util.InterfaceToJSON(data)
	}
	logger.Debug().
		Str("req_id", GetCtxReqId(ctx)).
		Interface("metadata", withContext(ctx)).
		Str("debug_type", string(debugType)).
		Msg(msg)
}
func Error(ctx context.Context, err error, data ...interface{}) {
	var msg string
	if len(data) > 0 {
		msg = util.InterfaceToJSON(data)
	}
	logger.Error().
		Str("req_id", GetCtxReqId(ctx)).
		Interface("metadata", withContext(ctx)).
		Err(err).
		Msg(msg)
}
