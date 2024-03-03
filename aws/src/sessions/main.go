package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	region    string = "ap-southeast-2"
	tableName string = "unisa-booking-sessions"
)

type Body struct {
	Message string `json:"message"`
	UUID    string `json:"hash"`
}

type Sessions struct {
	Groups []SessionGroup `json:"Sessions"`
}

type SessionGroup []Session

type Session struct {
	Date      string `json:"Date"`
	Details   string `json:"Details"`
	Available int    `json:"Available"`
}

func getClient() *dynamodb.DynamoDB {
	config := aws.NewConfig().WithRegion(region)

	session, err := session.NewSession()
	if err != nil {
		return nil
	}

	return dynamodb.New(session, config)
}

func createTable(client *dynamodb.DynamoDB) error {
	_, err := client.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("uuid"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("group"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("uuid"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("group"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	})

	return err
}

func checkTableExists() bool { return false }

func setSessions(client *dynamodb.DynamoDB) (Sessions, error) {
	wedsDetails := "Casual games at 15 minute intervals."
	thursDetails := "Casual games at 15 minute intervals."

	// set the session data.
	sessions := Sessions{
		Groups: []SessionGroup{
			{
				{
					Date:      "2024-03-06",
					Details:   wedsDetails,
					Available: 60,
				},
				{
					Date:      "2024-03-07",
					Details:   thursDetails,
					Available: 60,
				},
			},
			{
				{
					Date:      "2024-03-13",
					Details:   wedsDetails,
					Available: 60,
				},
				{
					Date:      "2024-03-14",
					Details:   thursDetails,
					Available: 60,
				},
			},
			{
				{
					Date:      "2024-03-20",
					Details:   wedsDetails,
					Available: 60,
				},
				{
					Date:      "2024-03-21",
					Details:   thursDetails,
					Available: 60,
				},
			},
			{
				{
					Date:      "2024-03-27",
					Details:   wedsDetails,
					Available: 60,
				},
				{
					Date:      "2024-03-28",
					Details:   thursDetails,
					Available: 60,
				},
			},
		},
	}

	// set the database.
	for groupIndex, group := range sessions.Groups {
		for sessionIndex, session := range group {
			id := fmt.Sprintf("%d-%d", groupIndex, sessionIndex) //uuid.NewString()

			_, err := client.PutItem(&dynamodb.PutItemInput{
				TableName: aws.String(tableName),
				Item: map[string]*dynamodb.AttributeValue{
					"uuid":      {S: aws.String(id)},
					"group":     {S: aws.String(fmt.Sprintf("%d", groupIndex))},
					"date":      {S: aws.String(session.Date)},
					"details":   {S: aws.String(session.Details)},
					"available": {N: aws.String(fmt.Sprintf("%d", session.Available))},
				},
			})
			if err != nil {
				return Sessions{}, err
			}
		}
	}

	return sessions, nil
}

func getSessions(client *dynamodb.DynamoDB) (Sessions, error) { return Sessions{}, nil }

func handleCors() {}
func handleGet()  {}

func respondWithStdErr(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 500,
	}, err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	client := getClient()

	// create the table.
	// err := createTable(client)
	// if err != nil {
	// 	return respondWithStdErr(err)
	// }

	sessions, err := setSessions(client)
	if err != nil {
		return respondWithStdErr(err)
	}

	fmt.Println(sessions)
	sessionJson, err := json.Marshal(sessions)
	if err != nil {
		return respondWithStdErr(err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(sessionJson),
		StatusCode: 201,
	}, nil
}

func main() {
	lambda.Start(handler)
}
