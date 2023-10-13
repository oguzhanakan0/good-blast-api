package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"oguzhanakan0/good-blast-api/api"
	"oguzhanakan0/good-blast-api/config"
	"oguzhanakan0/good-blast-api/structs"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestCreateUser(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.POST("/user", api.CreateUser)

	user := map[string]interface{}{
		"username": "TestUser#001",
		"country":  "TUR",
	}
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonValue))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetUser(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.POST("/user", api.CreateUser)
	r.GET("/user/:id", api.GetUser)

	user := map[string]interface{}{
		"username": "TestUser#002",
		"country":  "TUR",
	}
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonValue))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	// Get user
	var res map[string]string
	b, _ := io.ReadAll(w.Body)
	json.Unmarshal(b, &res)

	req, _ = http.NewRequest("GET", "/user/"+res["id"], bytes.NewBuffer([]byte{}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUsers(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.GET("/user/all", api.GetUsers)

	w := httptest.NewRecorder()
	// Get users
	req, _ := http.NewRequest("GET", "/user/all", bytes.NewBuffer([]byte{}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProgressUser(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.POST("/user", api.CreateUser)
	r.POST("/user/:id/progress", api.UpdateProgress)

	user := map[string]interface{}{
		"username": "TestUser#003",
		"country":  "TUR",
	}
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonValue))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	// Update progress
	var res map[string]string
	b, _ := io.ReadAll(w.Body)
	json.Unmarshal(b, &res)

	req, _ = http.NewRequest("POST", "/user/"+res["id"]+"/progress", bytes.NewBuffer([]byte{}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnterTournament(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.POST("/user", api.CreateUser)
	r.POST("/user/:id/tournament/:tournamentID/enter", api.EnterTournament)
	// Create a test user
	user := map[string]interface{}{
		"username":  "TestUser#004",
		"country":   "TUR",
		"gameLevel": 20,
		"coins":     10000,
	}
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var res map[string]string
	b, _ := io.ReadAll(w.Body)
	json.Unmarshal(b, &res)
	// Create a tournament
	to := structs.Tournament{
		ID:        "2000-01-01",
		Completed: false,
	}
	to.Put(db)
	// Enter tournament
	req, _ = http.NewRequest("POST", "/user/"+res["id"]+"/tournament/2000-01-01/enter", bytes.NewBuffer([]byte{}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if time.Now().UTC().Hour() >= config.TournamentEnterDeadline {
		assert.Equal(t, http.StatusForbidden, w.Code)
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestGetTournament(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.GET("/tournament/:id", api.GetTournament)
	// Create a tournament
	to := structs.Tournament{
		ID: "2000-01-01",
	}
	to.Put(db)
	// Get tournament
	req, _ := http.NewRequest("GET", "/tournament/2000-01-01", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTournaments(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	host := "http://localhost:8000"
	if os.Getenv("DYNAMODB_HOST") != "" {
		host = os.Getenv("DYNAMODB_HOST")
	}
	db := dynamodb.New(sess, aws.NewConfig().WithEndpoint(host))
	r := setupRouter()
	r.Use(dbMiddleware(db))
	r.GET("/tournament/all", api.GetTournaments)

	// Get tournaments
	req, _ := http.NewRequest("GET", "/tournament/all", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
