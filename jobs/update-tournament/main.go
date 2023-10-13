package main

import (
	"fmt"
	"oguzhanakan0/good-blast-api/structs"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	t := structs.Tournament{
		ID: time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02"),
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

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

	err := t.Fetch(db)
	if err != nil {
		panic(err)
	}

	if t.Completed {
		fmt.Printf("Results had already been calculated for tournament %s", t.ID)
		return
	}

	err = t.UpdateLeaderboards(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Results are calculated for tournament %s", t.ID)
}
