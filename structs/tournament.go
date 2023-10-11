package structs

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Tournament struct {
	ID           string        `json:"id"`
	Groups       []Group       `json:"groups"`
	Leaderboards []Leaderboard `json:"leaderboards"`
}

func (t *Tournament) Fetch(db *dynamodb.DynamoDB) error {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("tournament"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(t.ID)},
		},
	})

	if out.Item == nil {
		panic("Not found")
	}

	dynamodbattribute.UnmarshalMap(out.Item, &t)

	if err != nil {
		panic(err)
	}
	return nil
}

func (t *Tournament) AddUser(db *dynamodb.DynamoDB, u User) (int, error) {
	groupID := len(t.Groups) - 1
	group := t.Groups[groupID]

	if len(group.Users) < 3 {
		group.Users = append(group.Users, u.ID)
		t.Groups[groupID] = group
	} else {
		groupID += 1
		t.Groups = append(
			t.Groups,
			Group{ID: groupID, Users: []string{u.ID}},
		)
	}

	av, _ := dynamodbattribute.MarshalList(t.Groups)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("tournament"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(t.ID)},
		},
		UpdateExpression: aws.String("SET groups = :groups"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":groups": {L: av},
		},
	})
	if err != nil {
		return -1, errors.New("Cannot update the tournament.")
	}
	_ = out
	return groupID, nil
}

type Group struct {
	ID    int      `json:"id"`
	Users []string `json:"users"`
}

type Leaderboard struct {
	CountryCode string `json:"countryCode"`
	Users       []User `json:"users"`
}
