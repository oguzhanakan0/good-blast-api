package api

import (
	"fmt"
	"net/http"
	"oguzhanakan0/good-blast-api/config"
	"oguzhanakan0/good-blast-api/structs"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Creates a user in database.
func CreateUser(c *gin.Context) {
	// Instantiate a user
	user := structs.User{
		ID:    (uuid.New()).String(),
		Level: config.UserStartLevel,
		Coins: config.UserStartCoin,
	}

	// Parse JSON from request body
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Validator(s)
	if len(user.Username) < config.UsernameMinLength {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Username must contain at least 3 characters."})
		return
	}

	// Create user in database
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	out, err := user.Put(db)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	_ = out
	c.IndentedJSON(http.StatusCreated, user)
}

// Updates user information when they level up.
func UpdateProgress(c *gin.Context) {
	// Fetch user
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, err.Error())
		return
	}

	// Update progress (eg level up)
	tournamentID := time.Now().UTC().Format("2006-01-02")
	out, err := user.LevelUp(db, tournamentID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	_ = out

	c.Status(http.StatusOK)
}

// Returns a given user.
func GetUser(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

// Returns all the users in database.
func GetUsers(c *gin.Context) {
	// Scan database for all users
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	out, err := db.Scan(&dynamodb.ScanInput{TableName: aws.String("user")})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Populate an array with scan result
	var users []structs.User
	for _, e := range out.Items {
		var user structs.User
		dynamodbattribute.UnmarshalMap(e, &user)
		users = append(users, user)
	}
	c.IndentedJSON(http.StatusOK, users)
}

// Returns a given tournament.
func GetTournament(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	tournament := structs.Tournament{ID: c.Param("id")}
	err := tournament.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, tournament)
}

// Returns all tournaments.
func GetTournaments(c *gin.Context) {
	// Scan database for all tournaments
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	out, err := db.Scan(&dynamodb.ScanInput{TableName: aws.String("tournament")})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Populate an array with scan result
	var tournaments []structs.Tournament
	for _, e := range out.Items {
		var tournament structs.Tournament
		dynamodbattribute.UnmarshalMap(e, &tournament)
		tournaments = append(tournaments, tournament)
	}
	c.IndentedJSON(http.StatusOK, tournaments)
}

// Tries to add the user to today's tournament. If user passes all checks, they are added to the tournament.
func EnterTournament(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	// Fetch user
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	t := structs.Tournament{ID: c.Param("tournamentID")}
	err = t.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	// Check if user can enter the tournament
	if yes, err := user.CanEnterTournament(t); !yes {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	// Add user to the tournament
	err = user.EnterTournament(db, t)
	if err != nil {
		panic(err)
	}
	c.Status(http.StatusOK)
}

// Returns a given group.
func GetGroup(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	groupID, err := strconv.Atoi(c.Param("groupID"))
	group := structs.Group{TournamentID: c.Param("tournamentID"), GroupID: groupID}
	err = group.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, group)
}

// Returns all groups.
func GetGroups(c *gin.Context) {
	// Scan database for all groups
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	out, err := db.Scan(&dynamodb.ScanInput{TableName: aws.String("group")})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Populate an array with scan result
	var groups []structs.Group
	for _, e := range out.Items {
		var group structs.Group
		dynamodbattribute.UnmarshalMap(e, &group)
		groups = append(groups, group)
	}
	c.IndentedJSON(http.StatusOK, groups)
}

// Returns a given tournament and country leaderboard.
func GetLeaderboard(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	tournament := structs.Tournament{ID: c.Param("id")}
	err := tournament.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if !tournament.Completed {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "This tournament has not been completed yet."})
		return
	}

	board, ok := tournament.Leaderboards[c.Param("countryCode")]
	if !ok {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Country %s is not found in the leaderboards.", c.Param("countryCode"))})
		return
	}

	c.IndentedJSON(http.StatusOK, board)
}

// Returns a given user's group leaderboard.
func GetUserLeaderboard(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	// Get user
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	board, err := getUserLeaderboard(db, user, c.Param("tournamentID"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	_ = user
	c.IndentedJSON(http.StatusOK, board)
}

func ClaimReward(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	// Get the user
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	// Check if user has already claimed a reward
	if user.Tournaments[c.Param("tournamentID")].RewardClaimed {
		c.IndentedJSON(http.StatusAlreadyReported, gin.H{"message": "Reward is already claimed before."})
		return
	}
	// Get leaderboard for user's group
	board, err := getUserLeaderboard(db, user, c.Param("tournamentID"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	// Decide the amount
	var amount int
	for i, ur := range board {
		if i > 10 {
			break
		}
		if ur.UserID == user.ID {
			switch i {
			case 0:
				amount = config.TournamentReward1
			case 1:
				amount = config.TournamentReward2
			case 2:
				amount = config.TournamentReward3
			default:
				amount = config.TournamentRewardDefault
			}
		}
	}
	// Claim reward if amount is greater than zero
	if amount > 0 {
		err := user.ClaimReward(db, amount, c.Param("tournamentID"))
		if err != nil {
			panic(err)
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Claimed %d coins!", amount)})
		return
	}
	_ = board
	c.IndentedJSON(http.StatusNotModified, gin.H{"message": "No reward earned in this tournament :("})
}

func getUserLeaderboard(db *dynamodb.DynamoDB, user structs.User, tournamentID string) ([]structs.UserTournamentRecord, error) {
	var players []structs.UserTournamentRecord
	// Get group
	group := structs.Group{
		TournamentID: tournamentID,
		GroupID:      user.Tournaments[tournamentID].GroupID,
	}
	err := group.Fetch(db)
	if err != nil {
		return players, err
	}
	players = group.Players
	sort.Slice(players, func(i, j int) bool { return players[i].Score > players[j].Score })
	return players, nil
}
