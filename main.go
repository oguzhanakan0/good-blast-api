package main

import (
	"oguzhanakan0/good-blast-api/api"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func dbMiddleware(db *dynamodb.DynamoDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func main() {
	router := gin.Default()
	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	var db *dynamodb.DynamoDB
	// Set to local DynamoDB if not in release
	if gin.Mode() == "release" {
		db = dynamodb.New(sess)
	} else {
		host := "http://localhost:8000"
		if os.Getenv("DYNAMODB_HOST") != "" {
			host = os.Getenv("DYNAMODB_HOST")
		}
		db = dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	}
	router.Use(dbMiddleware(db))
	// User
	router.POST("/user", api.CreateUser)                                         //
	router.GET("/user/:id", api.GetUser)                                         //
	router.GET("/user/all", api.GetUsers)                                        //
	router.POST("/user/:id/progress", api.UpdateProgress)                        //
	router.POST("/user/:id/tournament/:tournamentID/enter", api.EnterTournament) //
	router.GET("/user/:id/tournament/:tournamentID/leaderboard", api.GetUserLeaderboard)
	router.POST("/user/:id/tournament/:tournamentID/claim-reward", api.ClaimReward)
	// Tournament
	router.GET("/tournament/:id", api.GetTournament)  //
	router.GET("/tournament/all", api.GetTournaments) //
	router.GET("/tournament/:id/leaderboard/:countryCode", api.GetLeaderboard)
	// Group
	router.GET("/group/:tournamentID/:groupID", api.GetGroup)
	router.GET("/group/all", api.GetGroups)

	router.Run(":8080")
}
