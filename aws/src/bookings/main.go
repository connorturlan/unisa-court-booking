package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	region    string = "ap-southeast-2"
	tableName string = "unisa-booking-register"
)

type BookingPullRequest struct {
	UID string `json:"uid"`
}

type BookingItem struct {
	UID      string   `json:"uid"`
	Sessions []string `json:"sessions"`
}

func getClient() *dynamodb.DynamoDB {
	config := aws.NewConfig().WithRegion(region).WithEndpoint("http://dynamodb-local:8000")

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
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("uid"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
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

func pushItem(client *dynamodb.DynamoDB, item BookingItem) error {
	// marshall the session so they can be sent.
	sessions, err := json.Marshal(item.Sessions)
	if err != nil {
		return err
	}

	// put the items into the table.
	_, err = client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"uid":      {S: aws.String(item.UID)},
			"sessions": {S: aws.String(string(sessions))},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func pullItem(client *dynamodb.DynamoDB, uid string) (BookingItem, error) {
	maybeItem, err := client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"uid": {S: aws.String(uid)},
		},
	})
	if err != nil {
		return BookingItem{}, err
	}

	if maybeItem.Item == nil {
		return BookingItem{}, nil
	}

	maybeSessions := *maybeItem.Item["sessions"].S
	var session []string
	err = json.Unmarshal([]byte(maybeSessions), &session)
	if err != nil {
		return BookingItem{}, err
	}

	booking := BookingItem{
		UID:      uid,
		Sessions: session,
	}

	return booking, nil
}

func respondWithStdErr(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 500,
	}, err
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
	fmt.Println("checking table exists.")
	if !checkTableExists(client) {
		createTable(client)
	}

	// process the body.
	fmt.Println("unmarshalling body.")
	var bookingRequestJson BookingPullRequest
	err := json.Unmarshal([]byte(request.Body), &bookingRequestJson)
	if err != nil {
		return respondWithStdErr(err)
	}

	// get the item.
	fmt.Println("pulling booking data.")
	booking, err := pullItem(client, bookingRequestJson.UID)
	if err != nil {
		return respondWithStdErr(err)
	}

	if booking.UID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "*",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
			},
		}, nil
	}

	// marshall the booking ready to send.
	fmt.Println("marshalling response.")
	bookingResponseJson, err := json.Marshal(booking)
	if err != nil {
		return respondWithStdErr(err)
	}

	// return.
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bookingResponseJson),
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
		},
	}, nil
}

func handlePost(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// create the dynamo client.
	client := getClient()

	// validate the table exists.
	fmt.Println("checking table exists.")
	if !checkTableExists(client) {
		createTable(client)
	}

	// process the body.
	fmt.Println("unmarshalling body.")
	var sessionsJson BookingItem
	err := json.Unmarshal([]byte(request.Body), &sessionsJson)
	if err != nil {
		return respondWithStdErr(err)
	}

	// push the body to the database.
	fmt.Println("pushing booking.")
	pushItem(client, sessionsJson)

	// return.
	return events.APIGatewayProxyResponse{
		StatusCode: 202,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
		},
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case http.MethodOptions:
		return handleCors(request)
	case http.MethodPost:
		return handlePost(request)
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
