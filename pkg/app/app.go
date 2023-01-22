package app

// https://github.com/gofiber/jwt
// https://github.com/gofiber/fiber
// https://github.com/awslabs/aws-lambda-go-api-proxy

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options for the app instance
type Option func(c *Config)

// Config for the app instance
type Config struct {
	log         *zerolog.Logger
	fiberLambda *fiberadapter.FiberLambda
	app         *fiber.App
	db          *database.DDB
}

// New creates a new mastoclinet instance
func New(opts ...Option) (*Config, error) {
	cfg := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	// set up logger if not provided
	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		cfg.log = &log
	}

	// Fail if no database is provided
	if cfg.db == nil {
		return nil, &NoDB{}
	}

	// Set up Fiber
	cfg.app = fiber.New()
	cfg.appSetup()
	cfg.fiberLambda = fiberadapter.New(cfg.app)

	return cfg, nil
}

// WithDB sets the database for the app instance
func WithDB(db *database.DDB) Option {
	return func(cfg *Config) {
		cfg.db = db
	}
}

// WithLogger sets the logger for the app instance
func WithLogger(log *zerolog.Logger) Option {
	return func(cfg *Config) {
		cfg.log = log
	}
}

// appSetup sets up the Fiber app
func (cfg *Config) appSetup() error {
	// Set up the "/" route
	cfg.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Add non-auth routes
	cfg.app.Get("/auth/callback", cfg.authCallback)
	cfg.app.Get("/auth/login", cfg.authLogin)

	// Install JWT Middleware
	// All following routes require a valid JWT
	if privateKey, err := cfg.getRSAPrivateKey(); err != nil {
		return err
	} else {
		cfg.app.Use(jwtware.New(jwtware.Config{
			SigningMethod: "RS256",
			SigningKey:    privateKey.Public(),
		}))
	}

	// Add auth routes
	cfg.app.Get("/auth/verify", cfg.authVerify)

	return nil
}

// getRSAPrivateKey gets the RSA private key from the database
func (cfg *Config) getRSAPrivateKey() (*rsa.PrivateKey, error) {
	// Get the RSA private key from the database
	jwtSingingKeyEncoded, err := cfg.db.GetConfig("jwt_signing_key")
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("function", "app::getRSAPrivateKey()::cfg.db.GetConfig('jwt_signing_key')").
			Str("errRef", guid.String()).
			Msg("Error getting jwt_signing_key from database")
		return nil, errors.New(guid.String() + ": Error to getting jwt_signing_key from database")
	}

	if jwtSingingKeyEncoded == nil {
		guid := xid.New()
		log.Error().
			Str("function", "app::getRSAPrivateKey()::cfg.db.GetConfig('jwt_signing_key')").
			Str("errRef", guid.String()).
			Msg("Error get jwt_signing_key from database")
		return nil, errors.New(guid.String() + ": Error to get jwt_signing_key from database")
	}

	// Decode the PEM formatted RSA private signing key
	block, _ := pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		guid := xid.New()
		log.Error().
			Str("function", "app::getRSAPrivateKey()::pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))").
			Str("errRef", guid.String()).
			Msg("Unable to decode jwt_signing_key PEM from database")
		return nil, errors.New(guid.String() + ": Unable to decode jwt_signing_key PEM from database")
	}

	// Parse the PEM encoded RSA private signing key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("function", "app::getRSAPrivateKey()::x509.ParsePKCS1PrivateKey(block.Bytes)").
			Str("errRef", guid.String()).
			Msg("Unable to parse PEM to RSA private key")
		return nil, errors.New(guid.String() + ": Unable to parse PEM to RSA private key")
	}

	return privateKey, nil
}

// LambdaHandler is the entry point for the Lambda function
func (cfg *Config) LambdaHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return cfg.fiberLambda.ProxyWithContextV2(ctx, req)
}
