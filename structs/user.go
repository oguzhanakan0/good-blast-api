package structs

import (
	"errors"
	"strconv"

	"oguzhanakan0/good-blast-api/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type User struct {
	ID          string                           `json:"id"`
	Username    string                           `json:"username"`
	Level       int                              `json:"gameLevel"`
	Coins       int                              `json:"coins"`
	Tournaments map[string]UserTournamentDetails `json:"tournaments"`
	Country     string                           `json:"country"`
}

type UserTournamentDetails struct {
	GroupID       int  `json:"groupID"`
	RewardClaimed bool `json:"rewardClaimed"`
}

type UserTournamentRecord struct {
	UserID  string `json:"userID"`
	Score   int    `json:"score"`
	Country string `json:"country"`
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

	err = dynamodbattribute.UnmarshalMap(out.Item, &u)

	if err != nil {
		return errors.New("Cannot parse the user.")
	}

	if u.Tournaments == nil {
		u.Tournaments = map[string]UserTournamentDetails{}
	}

	return nil
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

func (u *User) CanEnterTournament(t Tournament) (bool, error) {
	if t.Completed {
		return false, errors.New("This tournament has already been completed.")
	} else if _, alreadyIn := u.Tournaments[t.ID]; alreadyIn {
		return false, errors.New("User is already in the tournament.")
	} else if u.Coins < config.TournamentCost {
		return false, errors.New("Insufficient funds.")
	} else if u.Level < config.TournamentMinLevel {
		return false, errors.New(fmt.Sprintf("User must be above level %d.", config.TournamentMinLevel))
	} else if time.Now().UTC().Hour() >= config.TournamentEnterDeadline {
		return false, errors.New("Cannot enter the tournament in the afternoon.")
	}
	return true, nil
}

func (u *User) LevelUp(db *dynamodb.DynamoDB, tournamentID string) (*dynamodb.UpdateItemOutput, error) {
	// Update user level and coins
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
		UpdateExpression: aws.String("SET gameLevel = :gameLevel, coins = :coins"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gameLevel": {N: aws.String(strconv.Itoa(u.Level + config.ProgressLevelReward))},
			":coins":     {N: aws.String(strconv.Itoa(u.Coins + config.ProgressCoinReward))},
		},
	})
	// Update tournament score if the user is participating
	err = u.UpdateTournamentScore(db, tournamentID)
	return out, err
}

func (u *User) UpdateTournamentScore(db *dynamodb.DynamoDB, tournamentID string) error {
	// Update tournament score if the user is participating
	var err error
	if details, ok := u.Tournaments[tournamentID]; ok {
		group := Group{TournamentID: tournamentID, GroupID: details.GroupID}
		err := group.Fetch(db)
		if err != nil {
			return err
		}
		err = group.UpdateScore(db, u)
	}
	return err
}

func (u *User) ClaimReward(db *dynamodb.DynamoDB, amount int, tournamentID string) error {
	details := u.Tournaments[tournamentID]
	details.RewardClaimed = true
	u.Tournaments[tournamentID] = details
	av, _ := dynamodbattribute.MarshalMap(u.Tournaments)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
		UpdateExpression: aws.String("SET coins = :coins, tournaments = :tournaments"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tournaments": {M: av},
			":coins":       {N: aws.String(strconv.Itoa(u.Coins + amount))},
		},
	})
	_ = out
	return err
}

func (u *User) EnterTournament(db *dynamodb.DynamoDB, tournament Tournament) error {
	// Update group
	group, err := tournament.FetchLastGroup(db)
	if err != nil {
		panic(err)
	}
	// If there is no empty seat, create a new group and put the user in
	if group.Players == nil || len(group.Players) > config.GroupMaxLength {
		group = Group{
			TournamentID: tournament.ID,
			GroupID:      group.GroupID + 1,
			Players:      []UserTournamentRecord{{UserID: u.ID, Score: 0, Country: u.Country}},
		}
		out, err := group.Put(db)
		if err != nil {
			panic(err)
		}
		_ = out
	} else { // If there is an empty seat, put the user in that group
		err = group.AddUser(db, u)
		if err != nil {
			panic(err)
		}
	}
	// Update user model
	u.Tournaments[tournament.ID] = UserTournamentDetails{GroupID: group.GroupID, RewardClaimed: false}
	av, _ := dynamodbattribute.MarshalMap(u.Tournaments)
	out, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("user"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(u.ID)},
		},
		UpdateExpression: aws.String("SET tournaments = :tournaments, coins = :coins"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tournaments": {M: av},
			":coins":       {N: aws.String(strconv.Itoa(u.Coins - config.TournamentCost))},
		},
	})
	_ = out
	return err
}
