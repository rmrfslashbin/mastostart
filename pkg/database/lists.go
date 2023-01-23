package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// PutAppCredentials stores an app credentials item in the database.
func (config *DDB) PutList(list *List) error {
	item, err := attributevalue.MarshalMap(list)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(config.tableLists),
		Item:      item,
	}
	_, err = config.db.PutItem(context.TODO(), input)
	return err
}

func (config *DDB) PutAccountsInList(listMember *ListMember) error {
	input := &dynamodb.BatchWriteItemInput{}
	type entry struct {
		ListID string
		UserID string
	}
	for _, userID := range listMember.UserIDs {
		item, err := attributevalue.MarshalMap(entry{
			ListID: listMember.ListID,
			UserID: userID,
		})
		if err != nil {
			return err
		}
		input.RequestItems[config.tableAccountsInList] = append(
			input.RequestItems[config.tableAccountsInList],
			types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: item,
				},
			})
	}
	_, err := config.db.BatchWriteItem(context.TODO(), input)
	return err
}
