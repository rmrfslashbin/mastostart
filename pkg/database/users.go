package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DeleteUserCredentials deletes the user credentials from the database
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

// GetUserCredentials retrieves the user credentials from the database
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

// PutUserCredentials stores the user credentials in the database
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
