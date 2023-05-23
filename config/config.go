package config

import (
	"flag"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"sync"
)

type config struct {
	Key                  string `mapstructure:"KEY"`
	DBURI                string `mapstructure:"DATABASE_URI"`
	RunAddress           string `mapstructure:"RUN_ADDRESS"`
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"`
}

var instance *config
var once sync.Once

func GetConfig() *config {
	once.Do(func() {
		instance = &config{}
		// Order matters if we want to prioritize ENV over flags
		instance.readServerFlags()
		instance.readOs()
		log.Debug().Interface("config", instance).Msg("Server started with configs")
	})
	return instance
}

// readOs reads config from environment variables
// This func will replace Config parameters if any presented in os environment vars
func (config *config) readOs() {
	// load config from environment variables
	v := viper.New()
	v.AutomaticEnv()
	if v.Get("KEY") != nil {
		config.Key = v.GetString("ADDRESS")
	}
	if v.Get("DATABASE_URI") != nil {
		config.DBURI = v.GetString("DATABASE_URI")
	}
	if v.Get("RUN_ADDRESS") != nil {
		config.RunAddress = v.GetString("RUN_ADDRESS")
	}
	if v.Get("ACCRUAL_SYSTEM_ADDRESS") != nil {
		config.AccrualSystemAddress = v.GetString("ACCRUAL_SYSTEM_ADDRESS")
	}
}

// readServerFlags reads config from flags Run this first
func (config *config) readServerFlags() {
	// read flags
	appFlags := flag.NewFlagSet("go-metric-collector", flag.ContinueOnError)
	defaultURI := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	appFlags.StringVar(&config.Key, "k", "praktikum?sslmode=disable", "Key to use for authentication")
	appFlags.StringVar(&config.DBURI, "d", defaultURI, "Database URI")
	appFlags.StringVar(&config.RunAddress, "a", ":8080", "Run address")
	appFlags.StringVar(&config.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	err := appFlags.Parse(os.Args[1:])
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse flags")
	}
}
