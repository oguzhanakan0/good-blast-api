package api

import (
	"net/http"
	"oguzhanakan0/good-blast-api/structs"
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
		Level: 1,
		Coins: 1000,
	}

	// Parse JSON from request body
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Cannot parse the input."})
		return
	}

	// Validator(s)
	if len(user.Username) < 3 {
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
	out, err := user.LevelUp(db)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
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
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
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
	c.IndentedJSON(http.StatusCreated, users)
}

// Returns a given tournament.
func GetTournament(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	tournament := structs.Tournament{ID: c.Param("id")}
	err := tournament.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
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
	c.IndentedJSON(http.StatusCreated, tournaments)
}

// Tries to add the user to today's tournament.
// If user passes all checks, they are added to the tournament.
// There are two database updates:
// 1. User's `tournaments` attribute is updated to reflect the addition.
// We keep `groupID` and `score` on the user side.
// 2. Tournament's `groups` attribute is updated to reflect the addition.
// We keep a list of `userID`s for each group.
func EnterTournament(c *gin.Context) {
	db, _ := c.MustGet("db").(*dynamodb.DynamoDB)
	// Fetch the user
	user := structs.User{ID: c.Param("id")}
	err := user.Fetch(db)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, err.Error())
		return
	}

	tournament := structs.Tournament{ID: time.Now().UTC().Format("2006-01-02")}

	// Check if user can enter today's tournament
	if yes, err := user.CanEnterTournament(tournament.ID); !yes {
		c.IndentedJSON(http.StatusForbidden, err.Error())
		return
	}

	// Fetch the tournament
	err = tournament.Fetch(db)

	// Add user to the tournament, and tournament to the user
	groupID, err := tournament.AddUser(db, user)
	if err != nil {
		c.IndentedJSON(http.StatusConflict, err.Error())
		return
	}
	out, err := user.AddTournament(db, tournament.ID, groupID)
	_ = out
	c.IndentedJSON(http.StatusOK, tournament)
}
