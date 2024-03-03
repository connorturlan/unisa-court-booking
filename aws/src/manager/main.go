package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	region           string = "ap-southeast-2"
	bookingTableName string = "unisa-booking-register"
	sessionTableName string = "unisa-booking-sessions"
	sessionIndexName string = "GroupIndex"
)

type BookingItem struct {
	UID      string   `json:"uid"`
	Sessions []string `json:"sessions"`
}

type Sessions struct {
	Groups []*SessionGroup `json:"Sessions"`
}

type SessionGroup []*Session

type Session struct {
	UID            string `json:"uid"`
	Date           string `json:"Date"`
	Details        string `json:"Details"`
	TotalAvailable int    `json:"TotalAvailable"`
	Available      int    `json:"Available"`
}

func getClient() *dynamodb.DynamoDB {
	config := aws.NewConfig().WithRegion(region)

	session, err := session.NewSession()
	if err != nil {
		return nil
	}

	return dynamodb.New(session, config)
}

func setSessions(client *dynamodb.DynamoDB, sessions Sessions) (Sessions, error) {
	// set the database.
	for groupIndex, group := range sessions.Groups {
		for sessionIndex, session := range *group {
			id := fmt.Sprintf("%d-%d", groupIndex, sessionIndex) //uuid.NewString()

			_, err := client.PutItem(&dynamodb.PutItemInput{
				TableName: aws.String(sessionTableName),
				Item: map[string]*dynamodb.AttributeValue{
					"uid":            {S: aws.String(id)},
					"groupIndex":     {S: aws.String(fmt.Sprintf("%d", groupIndex))},
					"dateString":     {S: aws.String(session.Date)},
					"details":        {S: aws.String(session.Details)},
					"totalAvailable": {N: aws.String(fmt.Sprintf("%d", session.TotalAvailable))},
					"available":      {N: aws.String(fmt.Sprintf("%d", session.Available))},
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
			TableName:              aws.String(sessionTableName),
			IndexName:              aws.String(sessionIndexName),
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
			maybeUID, ok := sessionData["uid"]
			if !ok {
				return Sessions{}, nil
			}
			uid := *maybeUID.S

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

			maybeTotalAvailable, ok := sessionData["totalAvailable"]
			if !ok {
				return Sessions{}, nil
			}

			totalAvailable, err := strconv.Atoi(*maybeTotalAvailable.N)
			if err != nil {
				return Sessions{}, err
			}

			session := &Session{
				UID:            uid,
				Date:           date,
				Details:        details,
				TotalAvailable: totalAvailable,
				Available:      available,
			}

			group = append(group, session)
		}
		sessions.Groups = append(sessions.Groups, &group)

		groupIndex += 1
	}

	return sessions, nil
}

func getBookings(client *dynamodb.DynamoDB) ([]BookingItem, error) {
	allBookingsRaw, err := client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(bookingTableName),
	})
	if err != nil {
		return []BookingItem{}, err
	}

	allBookings := []BookingItem{}
	for _, bookingRaw := range allBookingsRaw.Items {
		sessionsRaw := *bookingRaw["sessions"].S

		var sessions []string
		err = json.Unmarshal([]byte(sessionsRaw), &sessions)
		if err != nil {
			return []BookingItem{}, err
		}

		booking := BookingItem{
			UID:      *bookingRaw["uid"].S,
			Sessions: sessions,
		}

		allBookings = append(allBookings, booking)
	}

	return allBookings, nil
}

func respondWithStdErr(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 500,
	}, err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// create client.
	client := getClient()

	// get all member bookings.
	bookings, err := getBookings(client)
	if err != nil {
		return respondWithStdErr(err)
	}
	fmt.Println(bookings)

	// aggregate counts.
	counts := map[string]int{}
	for _, booking := range bookings {
		for _, session := range booking.Sessions {
			if _, ok := counts[session]; !ok {
				counts[session] = 0
			}

			counts[session] += 1
		}
	}
	fmt.Println(counts)

	// get all the sessions.
	sessions, err := getSessions(client)
	if err != nil {
		return respondWithStdErr(err)
	}

	// update all session availability.
	for groupKey, count := range counts {
		for _, group := range sessions.Groups {
			for _, session := range *group {
				if groupKey == session.UID {
					session.Available = session.TotalAvailable - count
				}
			}
		}
	}

	// send the sessions.
	_, err = setSessions(client, sessions)
	if err != nil {
		return respondWithStdErr(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
