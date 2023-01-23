package database

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DDBOption is a function that configures the DDB struct
type DDBOption func(config *DDB)

// DDB is a struct that holds the DynamoDB client and table names
type DDB struct {
	db                   *dynamodb.Client
	profile              string
	region               string
	tablePrefix          string
	tableAccountsInList  string
	tableAppCredentials  string
	tableConfig          string
	tableLists           string
	tableUserCredentials string
}

// New returns a new DDB struct
func New(opts ...func(*DDB)) (*DDB, error) {
	cfg := &DDB{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the region from the environment if it's not set
	if cfg.region == "" {
		cfg.region = os.Getenv("AWS_REGION")
	}

	// Set the table prefix if it's not set
	if cfg.tablePrefix == "" {
		cfg.tablePrefix = "mastostart-"
	}

	// Set the table names
	cfg.tableAccountsInList = cfg.tablePrefix + "accounts-in-list"
	cfg.tableAppCredentials = cfg.tablePrefix + "app-credentials"
	cfg.tableConfig = cfg.tablePrefix + "config"
	cfg.tableUserCredentials = cfg.tablePrefix + "user-credentials"
	cfg.tableLists = cfg.tablePrefix + "lists"

	// Config DynamoDB
	c, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = cfg.region

		// If a profile is set, use it
		if cfg.profile != "" {
			o.SharedConfigProfile = cfg.profile
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Create the DynamoDB client
	svc := dynamodb.NewFromConfig(c)
	cfg.db = svc

	return cfg, nil
}

// WithDDBProfile sets the AWS profile to use
func WithDDBProfile(profile string) func(*DDB) {
	return func(config *DDB) {
		config.profile = profile
	}
}

// WithDDBRegion sets the AWS region to use
func WithDDBRegion(region string) func(*DDB) {
	return func(config *DDB) {
		config.region = region
	}
}

// WIthDDBTablePrefix sets the table prefix to use
func WithDDBTablePrefix(prefix string) func(*DDB) {
	return func(config *DDB) {
		config.tablePrefix = prefix
	}
}
