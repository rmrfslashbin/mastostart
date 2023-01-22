package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DeleteConfig deletes a config item from the database.
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

// GetConfig retrieves a config item from the database.
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

// PutConfig stores a config item in the database.
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
