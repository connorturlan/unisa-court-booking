package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	region    string = "ap-southeast-2"
	tableName string = "unisa-booking-sessions"
	indexName string = "GroupIndex"
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
				AttributeName: aws.String("uid"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("groupIndex"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("uid"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("groupIndex"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GroupIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("groupIndex"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(1),
					WriteCapacityUnits: aws.Int64(1),
				},
			},
		},
	})

	return err
}

func checkTableExists(client *dynamodb.DynamoDB) bool {
	awsTables, err := client.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return false
	}

	tables := []string{}
	for _, table := range awsTables.TableNames {
		tables = append(tables, *table)
	}

	return slices.Contains(tables, tableName)
}

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
					"uid":        {S: aws.String(id)},
					"groupIndex": {S: aws.String(fmt.Sprintf("%d", groupIndex))},
					"dateString": {S: aws.String(session.Date)},
					"details":    {S: aws.String(session.Details)},
					"available":  {N: aws.String(fmt.Sprintf("%d", session.Available))},
				},
			})
			if err != nil {
				return Sessions{}, err
			}
		}
	}

	return sessions, nil
}

func getSessions(client *dynamodb.DynamoDB) (Sessions, error) {
	sessions := Sessions{}

	groupIndex := 0
	for {
		output, err := client.Query(&dynamodb.QueryInput{
			TableName:              aws.String(tableName),
			IndexName:              aws.String(indexName),
			KeyConditionExpression: aws.String("groupIndex = :groupIndex"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":groupIndex": {S: aws.String(fmt.Sprintf("%d", groupIndex))},
			},
		})
		if err != nil {
			return Sessions{}, err
		}

		if *output.Count <= 0 {
			break
		}

		group := SessionGroup{}
		for _, sessionData := range output.Items {
			maybeDate, ok := sessionData["dateString"]
			if !ok {
				return Sessions{}, nil
			}
			date := *maybeDate.S

			maybeDetails, ok := sessionData["details"]
			if !ok {
				return Sessions{}, nil
			}
			details := *maybeDetails.S

			maybeAvailable, ok := sessionData["available"]
			if !ok {
				return Sessions{}, nil
			}

			available, err := strconv.Atoi(*maybeAvailable.N)
			if err != nil {
				return Sessions{}, err
			}

			session := Session{
				Date:      date,
				Details:   details,
				Available: available,
			}

			group = append(group, session)
		}
		sessions.Groups = append(sessions.Groups, group)

		groupIndex += 1
	}

	return sessions, nil
}

func handleCors(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
		},
	}, nil
}

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// create the dynamo client.
	client := getClient()

	// validate the table exists.
	// if !checkTableExists(client) {
	// 	createTable(client)
	// }

	// set the session data.
	// sessions, err := setSessions(client)
	// if err != nil {
	// 	return respondWithStdErr(err)
	// }

	// get the session data.
	sessions, err := getSessions(client)
	if err != nil {
		return respondWithStdErr(err)
	}

	// marshall the session ready to send.
	sessionJson, err := json.Marshal(sessions)
	if err != nil {
		return respondWithStdErr(err)
	}

	// return.
	return events.APIGatewayProxyResponse{
		Body:       string(sessionJson),
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
		},
	}, nil
}

func respondWithStdErr(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 500,
	}, err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case http.MethodOptions:
		return handleCors(request)
	case http.MethodGet:
		return handleGet(request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}
