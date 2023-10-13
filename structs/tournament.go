package structs

import (
	"errors"
	"oguzhanakan0/good-blast-api/config"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Tournament struct {
	ID           string              `json:"id"`
	Leaderboards map[string][]string `json:"leaderboards"` // format: { countryCode: Leaderboard }
	Completed    bool                `json:"completed"`    // true if the tournament has ended and results are calculated
}

func (t *Tournament) Fetch(db *dynamodb.DynamoDB) error {
	out, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("tournament"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(t.ID)},
		},
	})

	if out.Item == nil {
		return errors.New("Tournament does not exist.")
	}

	dynamodbattribute.UnmarshalMap(out.Item, &t)

	if err != nil {
		return errors.New("Cannot parse the tournament.")
	}
	return nil
}

func (t *Tournament) FetchGroups(db *dynamodb.DynamoDB) ([]Group, error) {
	out, err := db.Query(&dynamodb.QueryInput{
		TableName: aws.String("group"),
		KeyConditions: map[string]*dynamodb.Condition{
			"tournamentID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(t.ID),
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}

	var group []Group

	if len(out.Items) == 0 {
		return group, nil
	}

	var groups []Group
	for _, e := range out.Items {
		var group Group
		dynamodbattribute.UnmarshalMap(e, &group)
		groups = append(groups, group)
	}

	return groups, nil
}

func (t *Tournament) FetchLastGroup(db *dynamodb.DynamoDB) (Group, error) {
	out, err := db.Query(&dynamodb.QueryInput{
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(1),
		TableName:        aws.String("group"),
		KeyConditions: map[string]*dynamodb.Condition{
			"tournamentID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(t.ID),
					},
				},
			},
		},
	})

	if err != nil || len(out.Items) > 1 {
		panic(err)
	}

	var group Group

	if len(out.Items) == 0 {
		return group, nil
	}

	err = dynamodbattribute.UnmarshalMap(out.Items[0], &group)
	if err != nil {
		panic(err)
	}

	return group, nil
}

func (t *Tournament) Put(db *dynamodb.DynamoDB) (*dynamodb.PutItemOutput, error) {
	av, err := dynamodbattribute.MarshalMap(t)
	if err != nil {
		return nil, errors.New("Cannot marshal the tournament.")
	}
	out, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("tournament"),
		Item:      av,
	})
	return out, err
}

func (t *Tournament) UpdateLeaderboards(db *dynamodb.DynamoDB) error {
	// Fetch all groups for this tournament
	groups, err := t.FetchGroups(db)
	if err != nil {
		panic(err)
	}

	// Merge users into one map
	var players []UserTournamentRecord
	for _, group := range groups {
		for _, v := range group.Players {
			players = append(players, v)
		}
	}

	// Calculate leaderboards
	sort.Slice(players, func(i, j int) bool { return players[i].Score > players[j].Score })
	countries := map[string]bool{}
	leaderboards := map[string][]string{}
	leaderboards["ALL"] = []string{}
	for _, p := range players {
		countries[p.Country] = true
		if len(leaderboards["ALL"]) < config.GlobalLeaderboardMaxLength {
			leaderboards["ALL"] = append(leaderboards["ALL"], p.UserID)
		}
	}
	for country := range countries {
		var board []string
		for _, p := range players {
			if p.Country == country {
				board = append(board, p.UserID)
			}
			if len(board) >= config.LocalLeaderboardMaxLength {
				break
			}
		}
		leaderboards[country] = board
	}
	av, _ := dynamodbattribute.MarshalMap(leaderboards)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("tournament"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(t.ID)},
		},
		UpdateExpression: aws.String("SET leaderboards = :leaderboards, completed = :completed"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":leaderboards": {M: av},
			":completed":    {BOOL: aws.Bool(true)},
		},
	})
	if err != nil {
		panic(err)
	}
	_ = out
	return nil
}
