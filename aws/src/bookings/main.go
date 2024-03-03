package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	region string = "ap-southeast-2"
)

type Body struct {
	Message string `json:"message"`
	UUID    string `json:"hash"`
}

func getClient() *dynamodb.DynamoDB {
	config := aws.NewConfig().WithRegion(region)

	session, err := session.NewSession()
	if err != nil {
		return nil
	}

	return dynamodb.New(session, config)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// sourceIP := request.RequestContext.Identity.SourceIP

	client := getClient()

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"uuid": {
				S: aws.String("f208a62c770121a58fe0f25bf00ebec1fe933f3947287f0342d5eafca368e63e"),
			},
		},
		TableName: aws.String("unisa-hashes"),
	}

	result, err := client.GetItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}

	maybeHash := result.Item["uuid"]

	body := Body{
		Message: "success",
		UUID:    *maybeHash.S,
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("booking: %s\n", body),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
