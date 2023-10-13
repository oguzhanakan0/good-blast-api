package structs

import (
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Group struct {
	TournamentID string                 `json:"tournamentID"`
	GroupID      int                    `json:"groupID"`
	Players      []UserTournamentRecord `json:"players"`
}

func (g *Group) Fetch(db *dynamodb.DynamoDB) error {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("group"),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentID": {S: aws.String(g.TournamentID)},
			"groupID":      {N: aws.String(strconv.Itoa(g.GroupID))},
		},
	})

	if out.Item == nil {
		return errors.New("Not found")
	}

	err = dynamodbattribute.UnmarshalMap(out.Item, &g)
	return err
}

func (g *Group) Put(db *dynamodb.DynamoDB) (*dynamodb.PutItemOutput, error) {
	av, err := dynamodbattribute.MarshalMap(g)
	if err != nil {
		return nil, errors.New("Cannot marshal group.")
	}
	out, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("group"),
		Item:      av,
	})
	return out, err
}

func (g *Group) AddUser(db *dynamodb.DynamoDB, u *User) error {
	ur := UserTournamentRecord{UserID: u.ID, Score: 0, Country: u.Country}
	if g.Players == nil {
		g.Players = []UserTournamentRecord{ur}
	} else {
		g.Players = append(g.Players, ur)
	}
	av, _ := dynamodbattribute.MarshalList(g.Players)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("group"),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentID": {S: aws.String(g.TournamentID)},
			"groupID":      {N: aws.String(strconv.Itoa(g.GroupID))},
		},
		UpdateExpression: aws.String("SET players = :players"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":players": {L: av},
		},
	})
	if err != nil {
		return errors.New("Cannot add user to the group.")
	}
	_ = out
	return nil
}

func (g *Group) UpdateScore(db *dynamodb.DynamoDB, u *User) error {
	players := []UserTournamentRecord{}
	for _, ur := range g.Players {
		if ur.UserID == u.ID {
			ur.Score++
		}
		players = append(players, ur)
	}
	av, _ := dynamodbattribute.MarshalList(players)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("group"),
		Key: map[string]*dynamodb.AttributeValue{
			"tournamentID": {S: aws.String(g.TournamentID)},
			"groupID":      {N: aws.String(strconv.Itoa(g.GroupID))},
		},
		UpdateExpression: aws.String("SET players = :players"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":players": {L: av},
		},
	})
	if err != nil {
		return errors.New("Cannot add user to the group.")
	}
	_ = out
	return nil
}
