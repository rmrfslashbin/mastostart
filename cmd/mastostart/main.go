package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// APP_NAME is the name of the application
	APP_NAME = "mastostart"

	// CONFIG_FILE is the name of the config file
	CONFIG_FILE = "config.json"
)

// Context is used to pass context/global configs to the commands
type Context struct {
	// log is the logger
	log *zerolog.Logger
}

// ConfigSetCmd sets a config value
type ConfigSetCmd struct {
	Key     string `name:"key" required:"" enum:"app_name,permit_instances,redirect_uri,website," help:"The key to set."`
	Value   string `name:"value" required:"" help:"The value to set."`
	Profile string `name:"profile" default:"default" help:"The profile to set the value for."`
	Region  string `name:"region" default:"us-east-1" help:"The region to set the value for."`
	Prefix  string `name:"prefix" default:"mastostart-" help:"The prefix for dynamodb table names."`
}

// Run is the entry point for the config set command
func (r *ConfigSetCmd) Run(ctx *Context) error {
	db, err := database.New(
		database.SetDDBProfile(r.Profile),
		database.SetDDBRegion(r.Region),
		database.SetDDBTablePrefix(r.Prefix),
	)
	if err != nil {
		return err
	}
	if err := db.PutConfig(&database.ConfigItem{
		ConfigKey:   r.Key,
		ConfigValue: r.Value,
	}); err != nil {
		return err
	}
	log.Info().
		Str("key", r.Key).
		Str("value", r.Value).
		Str("aws profile", r.Profile).
		Str("aws region", r.Region).
		Str("ddb table prefix", r.Prefix).
		Msg("config set")
	return nil
}

// ConfigGetCmd gets a config value
type ConfigGetCmd struct {
	Key     string `name:"key" required:"" group:"selectors" xor:"selectors" help:"The key to get."`
	All     bool   `name:"all" required:"" group:"selectors" xor:"selectors" help:"Get all keys."`
	Profile string `name:"profile" default:"default" help:"The profile to set the value for."`
	Region  string `name:"region" default:"us-east-1" help:"The region to set the value for."`
	Prefix  string `name:"prefix" default:"mastostart-" help:"The prefix for dynamodb table names."`
}

// Run is the entry point for the config get command
func (r *ConfigGetCmd) Run(ctx *Context) error {
	db, err := database.New(
		database.SetDDBProfile(r.Profile),
		database.SetDDBRegion(r.Region),
		database.SetDDBTablePrefix(r.Prefix),
	)
	if err != nil {
		return err
	}
	var values []string
	if r.All {
		values = []string{"app_name", "website", "redirect_uri"}
	} else {
		values = []string{r.Key}
	}
	for _, v := range values {
		item, err := db.GetConfig(v)
		if err != nil {
			return err
		}
		if item == nil {
			log.Info().
				Str("key", v).
				Str("value", "**NOT SET**").
				Msg("config get")
		} else {
			log.Info().
				Str("key", item.ConfigKey).
				Str("value", item.ConfigValue).
				Msg("config get")
		}
	}
	return nil
}

// ConfigMakeJWTKeyCmd makes a JWT key
type ConfigMakeJWTKey struct {
	Profile string `name:"profile" default:"default" help:"The profile to set the value for."`
	Region  string `name:"region" default:"us-east-1" help:"The region to set the value for."`
	Prefix  string `name:"prefix" default:"mastostart-" help:"The prefix for dynamodb table names."`
	Len     int    `name:"len" default:"256" help:"The length of the key to generate."`
	Confirm bool   `name:"confirm" required:"" help:"Confirm the action. This will overwrite an existing key."`
}

// Run is the entry point for the config make-jwt-key command
func (r *ConfigMakeJWTKey) Run(ctx *Context) error {
	if !r.Confirm {
		return fmt.Errorf("you must confirm the action by passing --confirm")
	}

	db, err := database.New(
		database.SetDDBProfile(r.Profile),
		database.SetDDBRegion(r.Region),
		database.SetDDBTablePrefix(r.Prefix),
	)
	if err != nil {
		return err
	}

	rng := rand.Reader
	privateKey, err := rsa.GenerateKey(rng, 2048)
	if err != nil {
		ctx.log.Error().Err(err).Msg("failed to generate private key")
	}

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	if err := db.PutConfig(&database.ConfigItem{
		ConfigKey:   "jwt_signing_key",
		ConfigValue: string(pemdata),
	}); err != nil {
		return err
	}
	log.Info().
		Str("key", "jwt_signing_key").
		Str("value", "key_not_shown").
		Str("aws profile", r.Profile).
		Str("aws region", r.Region).
		Str("ddb table prefix", r.Prefix).
		Msg("config set")
	return nil

}

// ConfigCmd is the main config command
type ConfigCmd struct {
	Set    ConfigSetCmd     `cmd:"" help:"Set a config value."`
	Get    ConfigGetCmd     `cmd:"" help:"Get a config value."`
	JWTKey ConfigMakeJWTKey `cmd:"" help:"Make a JWT. This is a destructive action and will overwrite an existing key."`
}

// CLI is the main CLI struct
type CLI struct {
	// Global flags/args
	LogLevel string `name:"loglevel" env:"LOGLEVEL" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`

	//Cfg CfgCmd `cmd:"" help:"Show Mastgraph config details."`
	Config ConfigCmd `cmd:"" help:"Manage the config."`
}

func main() {
	var err error

	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Parse the command line
	var cli CLI
	ctx := kong.Parse(&cli)

	// Set up the logger's log level
	// Default to info via the CLI args
	switch cli.LogLevel {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	// Log some start up stuff for debugging
	log.Debug().Msg("Starting up")
	log.Debug().Msg("config paths/files")

	// Call the Run() method of the selected parsed command.
	err = ctx.Run(&Context{log: &log})

	// FatalIfErrorf terminates with an error message if err != nil
	ctx.FatalIfErrorf(err)
}
