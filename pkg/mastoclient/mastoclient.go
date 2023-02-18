package mastoclient

import (
	"context"
	"os"
	"strings"

	"github.com/mattn/go-mastodon"
	"github.com/rs/zerolog"
)

// Options for the mastoclient query
type Option func(c *Config)

// Config for the mastoclient query
type Config struct {
	log          *zerolog.Logger
	instance     *string
	clientKey    *string
	clientSecret *string
	accessToken  *string
}

// NewConfig creates a new Config
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

	return cfg, nil
}

// WithToken sets the token to use
func WithAccessToken(accessToken *string) Option {
	return func(cfg *Config) {
		cfg.accessToken = accessToken
	}
}

// WithClientID sets the client ID to use
func WithClientkey(clientKey *string) Option {
	return func(cfg *Config) {
		cfg.clientKey = clientKey
	}
}

// WithClientSecret sets the client secret to use
func WithClientSecret(clientSecret *string) Option {
	return func(cfg *Config) {
		cfg.clientSecret = clientSecret
	}
}

// WithInstance sets the instance to use
func WithInstance(instance *string) Option {
	return func(cfg *Config) {
		cfg.instance = instance
	}
}

// WithLogger sets the logger to use
func WithLogger(log *zerolog.Logger) Option {
	return func(cfg *Config) {
		cfg.log = log
	}
}

// SetAccessToken sets the access token
func (cfg *Config) SetAccessToken(accessToken *string) {
	cfg.accessToken = accessToken
}

// SetClientKey sets the client key
func (cfg *Config) SetClientKey(clientKey *string) {
	cfg.clientKey = clientKey
}

// SetClientSecret sets the client secret
func (cfg *Config) SetClientSecret(clientSecret *string) {
	cfg.clientSecret = clientSecret
}

// SetInstance sets the instance
func (cfg *Config) SetInstance(instance *string) {
	cfg.instance = instance
}

// SetLogger sets the logger
func (cfg *Config) SetLogger(log *zerolog.Logger) {
	cfg.log = log
}

// prefight checks if the config is set up correctly and returns a mastodon client
func (cfg *Config) preflight() (*mastodon.Client, error) {
	// set up a new mastodon client config struct
	clientConfig := &mastodon.Config{}

	if cfg.instance != nil {
		clientConfig.Server = *cfg.instance
	}

	if cfg.clientKey != nil {
		clientConfig.ClientID = *cfg.clientKey
	}

	if cfg.clientSecret != nil {
		clientConfig.ClientSecret = *cfg.clientSecret
	}

	if cfg.accessToken != nil {
		clientConfig.AccessToken = *cfg.accessToken
	}

	// Set up Mastodon client
	client := mastodon.NewClient(clientConfig)

	return client, nil
}

func (cfg *Config) GetAccountsInList(listId *mastodon.ID) ([]*mastodon.Account, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	accounts, err := client.GetListAccounts(context.Background(), *listId)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// GetAuthTokenFromCode gets an auth token from an auth code
func (cfg *Config) GetAuthTokenFromCode(authCode *string, redirectURI *string) (*string, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	if err = client.AuthenticateToken(context.Background(), *authCode, *redirectURI); err != nil {
		return nil, err
	}

	return &client.Config.AccessToken, nil
}

func (cfg *Config) GetInstanceInfo() (*mastodon.Instance, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	return client.GetInstance(context.Background())
}

func (cfg *Config) GetInstanceStats() ([]*mastodon.WeeklyActivity, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	return client.GetInstanceActivity(context.Background())
}

// GetLastStatus gets the last status of a user
func (cfg *Config) GetLastStatus(id *mastodon.ID) (*mastodon.Status, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	statuses, err := client.GetAccountStatuses(context.Background(), *id, &mastodon.Pagination{Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(statuses) < 1 {
		return nil, nil
	}

	return statuses[0], nil
}

// GetUserByID gets a user by ID
func (cfg *Config) GetUserByID(id string) (*mastodon.Account, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}
	// Get user

	return client.GetAccount(context.Background(), mastodon.ID(id))
}

// Me gets the current user
func (cfg *Config) Me() (*mastodon.Account, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}
	// Get user
	return client.GetAccountCurrentUser(context.Background())
}

// MyLists gets the lists of the current user. Set listId to get a specific list or nil to get all lists.
func (cfg *Config) MyLists(listId *mastodon.ID) ([]*mastodon.List, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	if listId != nil {
		if list, err := client.GetList(context.Background(), *listId); err != nil {
			return nil, err
		} else {
			return []*mastodon.List{list}, nil
		}
	} else {
		return client.GetLists(context.Background())
	}
}

// Post a toot
func (cfg *Config) Post(toot *mastodon.Toot) (*mastodon.ID, error) {
	client, err := cfg.preflight()
	if err != nil {
		return nil, err
	}

	// Post the toot
	if status, err := client.PostStatus(context.Background(), toot); err != nil {
		return nil, err
	} else {
		return &status.ID, nil
	}
}

// RegisterAppInput is the input for RegisterApp
func RegisterApp(input *RegisterAppInput) (*mastodon.Application, error) {
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:       input.InstanceURL,
		ClientName:   input.ClientName,
		RedirectURIs: input.RedirectURI,
		Scopes:       strings.Join(input.Scopes, " "),
		Website:      input.Website,
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}
