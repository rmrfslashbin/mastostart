package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DeleteAppCredentials deletes an app credentials item from the database.
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

// GetAppCredentials retrieves an app credentials item from the database.
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

// PutAppCredentials stores an app credentials item in the database.
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
