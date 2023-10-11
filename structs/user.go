package structs

import (
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type User struct {
	ID          string                    `json:"id"`
	Username    string                    `json:"username"`
	Level       int                       `json:"gameLevel"`
	Coins       int                       `json:"coins"`
	Tournaments map[string]UserTournament `json:"tournaments"`
}

func (u *User) Put(db *dynamodb.DynamoDB) (*dynamodb.PutItemOutput, error) {
	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New("Cannot marshal the user.")
	}
	out, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("user"),
		Item:      av,
	})
	return out, err
}

func (u *User) Fetch(db *dynamodb.DynamoDB) error {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
	})

	if out.Item == nil {
		return errors.New("User does not exist.")
	}

	var user User
	dynamodbattribute.UnmarshalMap(out.Item, &user)

	if err != nil {
		return errors.New("Cannot parse the user.")
	}

	if user.Tournaments == nil {
		user.Tournaments = map[string]UserTournament{}
	}

	return nil
}

func (u *User) CanEnterTournament(tournamentID string) (bool, error) {
	// @TODO: Enable later
	// if _, alreadyIn := u.Tournaments[tournamentID]; alreadyIn {
	// 	return false, errors.New("User is already in the tournament.")
	// } else if u.Coins < 500 {
	// 	return false, errors.New("Insufficient funds.")
	// } else if u.Level < 10 {
	// 	return false, errors.New("User must be above level 10.")
	// } else if time.Now().UTC().Hour() >= 12 {
	// 	return false, errors.New("Cannot enter the tournament in the afternoon.")
	// }
	return true, nil
}

func (u *User) LevelUp(db *dynamodb.DynamoDB) (*dynamodb.UpdateItemOutput, error) {
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
		UpdateExpression: aws.String("SET gameLevel = :gameLevel, coins = :coins"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gameLevel": {N: aws.String(strconv.Itoa(u.Level + 1))},
			":coins":     {N: aws.String(strconv.Itoa(u.Coins + 100))},
		},
	})
	return out, err
}

func (u *User) AddTournament(db *dynamodb.DynamoDB, tournamentID string, groupID int) (*dynamodb.UpdateItemOutput, error) {
	u.Tournaments[tournamentID] = UserTournament{GroupID: groupID, Score: 0}
	av, _ := dynamodbattribute.MarshalMap(u.Tournaments)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
		UpdateExpression: aws.String("SET tournaments = :tournaments, coins = :coins"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tournaments": {M: av},
			":coins":       {N: aws.String(strconv.Itoa(u.Coins - 500))},
		},
	})
	return out, err
}

type UserTournament struct {
	GroupID int `json:"groupID"`
	Score   int `json:"score"`
}
