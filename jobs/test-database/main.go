package main

import (
	"fmt"
	"log"
	"math/rand"
	"oguzhanakan0/good-blast-api/structs"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	createTables()
	insertData()
}

func createTables() {
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

	tableNames := []string{"user", "tournament", "group"}
	for _, tableName := range tableNames {
		_, err := db.DeleteTable(&dynamodb.DeleteTableInput{
			TableName: aws.String(tableName)})
		if err != nil {
			fmt.Printf("Couldn't delete table %v: %v\n", tableName, err)
		} else {
			fmt.Printf("Deleted table %v.\n", tableName)
		}
	}

	userTableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("user"),
	}

	_, err := db.CreateTable(userTableInput)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
	}

	tournamentTableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("tournament"),
	}

	_, err = db.CreateTable(tournamentTableInput)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
	}

	groupTableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("tournamentID"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("groupID"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("tournamentID"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("groupID"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("group"),
	}

	_, err = db.CreateTable(groupTableInput)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
	}

	fmt.Println("Recreated all tables.")
}

func RandomUsername(n int) string {
	f := []string{"Meadow", "Aron", "Briana", "Jax", "Ayleen", "Zayn", "Vera", "Easton", "Sawyer",
		"Wilder", "Mikayla", "Wesson", "Emmalynn", "Devin", "Adalynn", "Larry", "Marianna", "Kieran", "Malaysia",
		"Deandre", "Remington", "Morgan", "Giavanna", "Miguel", "Juliana", "Cohen", "Rosalee", "Parker", "Mckenna", "Jeremiah"}
	l := []string{"Prince", "Hood", "Burke", "Church", "Cortez", "Burke", "Reed", "Estrada",
		"Rosales", "Owen", "Trejo", "Dickson", "McCarthy", "Jordan", "Walls", "Corona", "Bauer", "Singleton",
		"Winters", "Macias", "Ho", "Lim", "Ferguson", "Ferguson", "Wang", "Blankenship", "Patel", "Howell", "Howard", "Gallegos"}

	return f[rand.Intn(len(f))] + l[rand.Intn(len(l))] + "#" + strconv.Itoa(rand.Intn(100))
}

func insertData() {
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
	// Insert today's and yesterday's tournament
	t := structs.Tournament{
		ID: time.Now().UTC().Format("2006-01-02"),
	}
	out, err := t.Put(db)
	t = structs.Tournament{
		ID: time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02"),
	}
	out, err = t.Put(db)
	if err != nil {
		panic(err)
	}
	_ = out
	fmt.Printf("[%s] Inserted tournament\n", t.ID)

	// Insert users with random scores
	countries := []string{"TUR", "US"}
	fmt.Println("Inserting users...")
	for i := 0; i < 100; i++ {
		if j := (i % 10); j == 0 {
			fmt.Println(i)
		}
		u := structs.User{
			ID:          (uuid.New()).String(),
			Level:       rand.Intn(90) + 10,
			Coins:       (rand.Intn(90) + 10) * 100,
			Username:    RandomUsername(5),
			Country:     countries[rand.Intn(len(countries))],
			Tournaments: map[string]structs.UserTournamentDetails{},
		}
		// Enter tournament
		err := u.EnterTournament(db, t)
		if err != nil {
			panic(err)
		}
		// Level up randomly
		for k := 0; k < rand.Intn(5); k++ {
			u.LevelUp(db, t.ID)
		}
		out, err := u.Put(db)
		if err != nil {
			panic(err)
		}
		_ = out
	}

	// End tournament & calculate results
	err = t.UpdateLeaderboards(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[%s] Updated leaderboards\n", t.ID)
}
