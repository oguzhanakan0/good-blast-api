package main

import (
	"fmt"
	"oguzhanakan0/good-blast-api/structs"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func main() {
	t := structs.Tournament{
		ID:     time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02"),
		Groups: []structs.Group{{ID: 0, Users: []string{}}},
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	db := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(t)
	if err != nil {
		panic(err)
	}

	out, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("tournament"),
		Item:      av,
	})

	if err != nil {
		panic(err)
	}
	_ = out
	fmt.Printf("Inserted tournament for %s", t.ID)
}
