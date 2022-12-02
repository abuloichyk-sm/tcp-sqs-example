package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	qcIn := SqsQueueClient{}
	err := qcIn.Init(sess, aws.String("AlexTestQueueEngineIn.fifo"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	qcOut := SqsQueueClient{}
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

func ProcessMessage(qcIn *SqsQueueClient, qcOut *SqsQueueClient, message *sqs.Message) {

	engineRes := EngineResponse{Id: *message.Body, Message: fmt.Sprintf("%s processed", *message.Body)}
	log.Println(engineRes.Message)

	outMessageBytes, _ := json.Marshal(engineRes)
	outMessage := string(outMessageBytes)
	err := qcOut.SendMsg(&outMessage)
	if err == nil {
		qcIn.DeleteMessage(message.ReceiptHandle)
	}
}
