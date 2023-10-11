package main

import (
	"oguzhanakan0/good-blast-api/api"

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
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	router := gin.Default()
	router.Use(dbMiddleware(svc))
	router.GET("/user/all", api.GetUsers)
	router.POST("/user/new", api.CreateUser)
	router.GET("/user/:id", api.GetUser)
	router.POST("/user/:id/progress", api.UpdateProgress)
	router.POST("/user/:id/enter-tournament", api.EnterTournament)
	router.GET("/tournament/:id", api.GetTournament)
	router.GET("/tournament/all", api.GetTournaments)

	router.Run("localhost:8080")
}
