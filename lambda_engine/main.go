package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
)

type QueueRequest struct {
	//Records []QueueMessage `json:"Records"`
	Records []types.Message
}

type QueueMessage struct {
	MessageId string `json:"messageId"`
	Body      string `json:"body"`
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, qr *QueueRequest) (string, error) {
	fmt.Printf("HandleRequest %v\n", qr)

	res := "response"
	return SendToOutQueue(ctx, &res)
}

func SendToOutQueue(ctx context.Context, body *string) (string, error) {
	cfg, _ := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-central-1"))
	fmt.Printf("cfg %v\n", cfg)
	svc := sqs.NewFromConfig(cfg)

	gQInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String("BuloichykEngineOut"),
	}

	result, err := svc.GetQueueUrl(ctx, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return "error get queue url", nil
	}

	queueURL := result.QueueUrl

	sMInput := &sqs.SendMessageInput{
		MessageBody: body,
		QueueUrl:    queueURL,
	}

	svc.SendMessage(ctx, sMInput)

	return fmt.Sprintf("Hello! Err send message to queue '%v'", err), nil
}

// func ProcessMessage(qcIn *SqsQueueClient, qcOut *SqsQueueClient, message *sqs.Message) {
// 	//time measures
// 	start := time.Now()

// 	req := &EngineRequest{}
// 	err := json.Unmarshal([]byte(*message.Body), req)
// 	if err != nil {
// 		log.Printf("Error to json.Unmarshal '%s'", *message.Body)
// 		return
// 	}
// 	transaction, err := base64.StdEncoding.DecodeString(*req.B64Message)
// 	if err != nil {
// 		log.Printf("Error to base64 decode '%s'", *req.B64Message)
// 		return
// 	}

// 	//time measures
// 	logTime(start, "Message received", req.Id)

// 	//processing
// 	res := fmt.Sprintf("%s processed", transaction)

// 	//log
// 	//log.Printf("%s processed", transaction)

// 	// logged in message
// 	res = fmt.Sprintf("%s| Engine received time %s|", res, start.Format(time.StampMilli))

// 	b64res := base64.StdEncoding.EncodeToString([]byte(res))
// 	engineRes := EngineResponse{Id: req.Id, B64Message: &b64res}

// 	outMessageBytes, _ := json.Marshal(engineRes)
// 	outMessage := string(outMessageBytes)

// 	//time measures
// 	//logTime(start, "Response message marshaled", req.Id)

// 	err = qcOut.SendMsg(&outMessage, req.Id)

// 	//time measures
// 	logTime(start, "Total processing time", req.Id)

// 	if err == nil {
// 		qcIn.DeleteMessage(message.ReceiptHandle)
// 	}
// }

func logTime(start time.Time, m string, id *string) {
	elapsed := time.Since(start)
	requestNumber := strings.Split(*id, "_")[0]
	log.Printf("Request '%s'. %s - %s", requestNumber, m, elapsed)
}
