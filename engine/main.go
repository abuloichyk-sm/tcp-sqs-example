package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("Messages count %d", len(out.Messages))
		for _, m := range out.Messages {
			go ProcessMessage(&qcIn, &qcOut, m)
		}
	}
}

func ProcessMessage(qcIn *queueclient.SqsQueueClient, qcOut *queueclient.SqsQueueClient, message *sqs.Message) {

	engineRes := queueclient.EngineResponse{Id: *message.Body, Message: fmt.Sprintf("%s processed", *message.Body)}
	log.Println(engineRes.Message)

	outMessageBytes, _ := json.Marshal(engineRes)
	outMessage := string(outMessageBytes)
	deduplicaionId := fmt.Sprintf("%s_%d", *message.Body, time.Now().Unix())
	err := qcOut.SendMsg(&outMessage, &deduplicaionId)
	if err == nil {
		qcIn.DeleteMessage(message.ReceiptHandle)
	}
}
