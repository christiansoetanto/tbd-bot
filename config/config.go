package config

import (
	"os"
)

type RegisteredFeature map[string]bool
type AppConfig struct {
	DevMode                    bool
	FirebaseServiceAccountJson string
	TursoAuth                  string
	TursoName                  string
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
	guildCfg := buildGuildConfig()
	return Config{
		AppConfig:   appCfg,
		GuildConfig: guildCfg,
	}
}
