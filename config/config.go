package config

import (
	"github.com/christiansoetanto/tbd-bot/domain"
	"os"
)

type RegisteredFeature map[domain.FeatureKey]bool
type AppConfig struct {
	DevMode                    bool
	FirebaseServiceAccountJson string
}

type Config struct {
	GuildConfig map[GuildId]GuildConfig
	AppConfig   AppConfig
}

func Init(devMode bool) Config {
	appCfg := AppConfig{
		DevMode:                    devMode,
		FirebaseServiceAccountJson: os.Getenv("FIREBASE_CONFIG"),
	}
	guildCfg := make(map[GuildId]GuildConfig)
	if devMode {
		guildCfg[devGuild.GuildId] = devGuild
	} else {
		//TODO ganti ke SD
		guildCfg[devGuild.GuildId] = devGuild
	}
	return Config{
		AppConfig:   appCfg,
		GuildConfig: guildCfg,
	}
}
