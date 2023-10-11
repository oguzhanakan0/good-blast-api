package api

import (
	"oguzhanakan0/good-blast-api/structs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func fetchUserStruct(db *dynamodb.DynamoDB, id string) (structs.User, error) {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	})

	if out.Item == nil {
		panic("Not found")
	}

	var user structs.User
	dynamodbattribute.UnmarshalMap(out.Item, &user)

	if err != nil {
		panic(err)
	}

	if user.Tournaments == nil {
		user.Tournaments = map[string]structs.UserTournament{}
	}

	return user, nil
}

func fetchTournamentStruct(db *dynamodb.DynamoDB, id string) (structs.Tournament, error) {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("tournament"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	})

	if out.Item == nil {
		panic("Not found")
	}

	var tournament structs.Tournament
	dynamodbattribute.UnmarshalMap(out.Item, &tournament)

	if err != nil {
		panic(err)
	}

	return tournament, nil
}
