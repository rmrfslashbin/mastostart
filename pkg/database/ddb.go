package database

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type AppCredentials struct {
	/*
		{
			"id": "13",
			"name": "Test Application 8421",
			"website": "https://myapp.example",
			"redirect_uri": "http://localhost:8421",
			"client_id": "sample_client_id",
			"client_secret": "sample_client_secret",
			"vapid_key": "sample_vapid_key"
		}
	*/

	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Website      string `json:"website"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthURI      string `json:"vapid_key"`
}

type ConfigItem struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}

type UserCredentials struct {
	InstanceURL string `json:"instance_url"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	CreatedAt   string `json:"created_at"`
}

type DDBOption func(config *DDB)

type DDB struct {
	db                   *dynamodb.Client
	profile              string
	region               string
	tablePrefix          string
	tableAppCredentials  string
	tableConfig          string
	tableUserCredentials string
}

func New(opts ...func(*DDB)) (*DDB, error) {
	cfg := &DDB{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.region == "" {
		cfg.region = os.Getenv("AWS_REGION")
	}

	c, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = cfg.region

		if cfg.profile != "" {
			o.SharedConfigProfile = cfg.profile
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	svc := dynamodb.NewFromConfig(c)
	cfg.db = svc

	if cfg.tablePrefix == "" {
		cfg.tablePrefix = "mastostart-"
	}

	cfg.tableAppCredentials = cfg.tablePrefix + "app-credentials"
	cfg.tableConfig = cfg.tablePrefix + "config"
	cfg.tableUserCredentials = cfg.tablePrefix + "user-credentials"

	return cfg, nil
}

func SetDDBProfile(profile string) func(*DDB) {
	return func(config *DDB) {
		config.profile = profile
	}
}

func SetDDBRegion(region string) func(*DDB) {
	return func(config *DDB) {
		config.region = region
	}
}

func SetDDBTablePrefix(prefix string) func(*DDB) {
	return func(config *DDB) {
		config.tablePrefix = prefix
	}
}

func (config *DDB) GetAppCredentials(instance string) (*AppCredentials, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.tableAppCredentials),
		Key: map[string]types.AttributeValue{
			"InstanceURL": &types.AttributeValueMemberS{Value: instance},
		},
	}
	result, err := config.db.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	app := &AppCredentials{}
	err = attributevalue.UnmarshalMap(result.Item, app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (config *DDB) PutAppCredentials(app *AppCredentials) error {
	item, err := attributevalue.MarshalMap(app)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(config.tableAppCredentials),
		Item:      item,
	}
	_, err = config.db.PutItem(context.TODO(), input)
	return err
}

func (config *DDB) DeleteAppCredentials(instance string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.tableAppCredentials),
		Key: map[string]types.AttributeValue{
			"InstanceURL": &types.AttributeValueMemberS{Value: instance},
		},
	}
	_, err := config.db.DeleteItem(context.TODO(), input)
	return err
}

func (config *DDB) GetUserCredentials(instance string, userId string) (*UserCredentials, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.tableUserCredentials),
		Key: map[string]types.AttributeValue{
			"InstanceURL": &types.AttributeValueMemberS{Value: instance},
			"UserID":      &types.AttributeValueMemberS{Value: userId},
		},
	}
	result, err := config.db.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	user := &UserCredentials{}
	err = attributevalue.UnmarshalMap(result.Item, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (config *DDB) PutUserCredentials(cred *UserCredentials) error {
	item, err := attributevalue.MarshalMap(cred)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(config.tableUserCredentials),
		Item:      item,
	}
	_, err = config.db.PutItem(context.TODO(), input)
	return err
}

func (config *DDB) DeleteUserCredentials(instance string, userId string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.tableUserCredentials),
		Key: map[string]types.AttributeValue{
			"InstanceURL": &types.AttributeValueMemberS{Value: instance},
			"UserID":      &types.AttributeValueMemberS{Value: userId},
		},
	}
	_, err := config.db.DeleteItem(context.TODO(), input)
	return err
}

func (config *DDB) DeleteConfig(key string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.tableConfig),
		Key: map[string]types.AttributeValue{
			"ConfigKey": &types.AttributeValueMemberS{Value: key},
		},
	}
	_, err := config.db.DeleteItem(context.TODO(), input)
	return err
}

func (config *DDB) GetConfig(key string) (*ConfigItem, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.tableConfig),
		Key: map[string]types.AttributeValue{
			"ConfigKey": &types.AttributeValueMemberS{Value: key},
		},
	}
	result, err := config.db.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	item := &ConfigItem{}
	err = attributevalue.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (config *DDB) PutConfig(item *ConfigItem) error {
	m, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	//spew.Dump(config)
	input := &dynamodb.PutItemInput{
		TableName: aws.String(config.tableConfig),
		Item:      m,
	}
	_, err = config.db.PutItem(context.TODO(), input)
	return err
}
