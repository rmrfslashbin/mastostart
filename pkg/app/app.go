package app

// https://github.com/gofiber/jwt
// https://github.com/gofiber/fiber
// https://github.com/awslabs/aws-lambda-go-api-proxy

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type JWTClaims struct {
	AccessToken string `json:"access_token"`
	jwt.RegisteredClaims
}

type LoginUserOutput struct {
	UserCreds *database.UserCredentials
	AuthURI   *string
}

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

	if cfg.db == nil {
		return nil, &NoDB{}
	}

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

func (cfg *Config) appSetup() error {
	jwtSingingKeyEncoded, err := cfg.db.GetConfig("jwt_signing_key")
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("function", "app::appSetup()::cfg.db.GetConfig('jwt_signing_key')").
			Str("errRef", guid.String()).
			Msg("Error getting jwt_signing_key from database")
		return errors.New(guid.String() + ": Error to getting jwt_signing_key from database")
	}

	if jwtSingingKeyEncoded == nil {
		guid := xid.New()
		log.Error().
			Str("function", "app::appSetup()::cfg.db.GetConfig('jwt_signing_key')").
			Str("errRef", guid.String()).
			Msg("Error get jwt_signing_key from database")
		return errors.New(guid.String() + ": Error to get jwt_signing_key from database")
	}

	// Decode the PEM formatted RSA private signing key
	block, _ := pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		guid := xid.New()
		log.Error().
			Str("function", "app::appSetup()::pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))").
			Str("errRef", guid.String()).
			Msg("Unable to decode jwt_signing_key PEM from database")
		return errors.New(guid.String() + ": Unable to decode jwt_signing_key PEM from database")
	}

	// Parse the PEM encoded RSA private signing key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("function", "app::appSetup()::x509.ParsePKCS1PrivateKey(block.Bytes)").
			Str("errRef", guid.String()).
			Msg("Unable to parse PEM to RSA private key")
		return errors.New(guid.String() + ": Unable to parse PEM to RSA private key")
	}

	cfg.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	cfg.app.Get("/auth/callback", cfg.authCallback)
	cfg.app.Get("/auth/login", cfg.authLogin)

	// JWT Middleware
	cfg.app.Use(jwtware.New(jwtware.Config{
		SigningMethod: "RS256",
		SigningKey:    privateKey.Public(),
	}))

	cfg.app.Get("/auth/verify", cfg.authVerify)

	return nil
}

// LambdaHandler will deal with Fiber working with Lambda
func (cfg *Config) LambdaHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return cfg.fiberLambda.ProxyWithContextV2(ctx, req)
}
