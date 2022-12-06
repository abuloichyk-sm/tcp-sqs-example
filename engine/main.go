package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	queueclient "github.com/abuloichyk-sm/tcp-sqs-example/internal/queueclient"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", ""),
		Region:      aws.String("us-east-1")},
	))

	qcIn := queueclient.SqsQueueClient{}
	err := qcIn.Init(sess, aws.String("AlexTestQueueEngineIn.fifo"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	qcOut := queueclient.SqsQueueClient{}
	err = qcOut.Init(sess, aws.String("AlexTestQueueEngineOut.fifo"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine out queue client created")

	t := time.NewTicker(200 * time.Millisecond)
	for {
		<-t.C
		out, _ := qcIn.ReceiveMessages()
		if len(out.Messages) == 0 {
			log.Println("No messages")
			continue
		}
		log.Printf("Messages count %d, current time %s", len(out.Messages), time.Now())
		for _, m := range out.Messages {
			go ProcessMessage(&qcIn, &qcOut, m)
		}
	}
}

func ProcessMessage(qcIn *queueclient.SqsQueueClient, qcOut *queueclient.SqsQueueClient, message *sqs.Message) {
	//time measures
	start := time.Now()

	req := &queueclient.EngineRequest{}
	err := json.Unmarshal([]byte(*message.Body), req)
	if err != nil {
		log.Printf("Error to json.Unmarshal '%s'", *message.Body)
		return
	}
	transaction, err := base64.StdEncoding.DecodeString(*req.B64Message)
	if err != nil {
		log.Printf("Error to base64 decode '%s'", req.B64Message)
		return
	}

	//processing
	res := fmt.Sprintf("%s processed", transaction)
	log.Printf("%s processed", transaction)

	b64res := base64.StdEncoding.EncodeToString([]byte(res))
	engineRes := queueclient.EngineResponse{Id: req.Id, B64Message: &b64res}

	outMessageBytes, _ := json.Marshal(engineRes)
	outMessage := string(outMessageBytes)

	err = qcOut.SendMsg(&outMessage, req.Id)

	//time measures
	elapsed := time.Since(start)
	requestNumber := strings.Split(*req.Id, "_")[0]
	log.Printf("Request %s - %s", requestNumber, elapsed)

	if err == nil {
		qcIn.DeleteMessage(message.ReceiptHandle)
	}
}
