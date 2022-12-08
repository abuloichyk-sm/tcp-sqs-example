package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", "siliconmint"),
		Region:      aws.String("eu-central-1")},
	))

	qcIn := SqsQueueClient{}
	err := qcIn.Init(sess, aws.String("BuloichykEngineIn"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	qcOut := SqsQueueClient{}
	err = qcOut.Init(sess, aws.String("BuloichykEngineOut"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine out queue client created")

	t := time.NewTicker(100 * time.Millisecond)
	for {
		<-t.C
		out, _ := qcIn.ReceiveMessages()

		if len(out.Messages) == 0 {
			//debug output
			//log.Println("No messages")

			continue
		}

		log.Printf("Messages recived. Count %d", len(out.Messages))
		for _, m := range out.Messages {
			go ProcessMessage(&qcIn, &qcOut, m)
		}
	}
}

func ProcessMessage(qcIn *SqsQueueClient, qcOut *SqsQueueClient, message *sqs.Message) {
	//time measures
	start := time.Now()

	req := &EngineRequest{}
	err := json.Unmarshal([]byte(*message.Body), req)
	if err != nil {
		log.Printf("Error to json.Unmarshal '%s'", *message.Body)
		return
	}
	transaction, err := base64.StdEncoding.DecodeString(*req.B64Message)
	if err != nil {
		log.Printf("Error to base64 decode '%s'", *req.B64Message)
		return
	}

	//time measures
	logTime(start, "JSON parsed", req.Id)

	//processing
	res := fmt.Sprintf("%s processed", transaction)
	log.Printf("%s processed", transaction)

	b64res := base64.StdEncoding.EncodeToString([]byte(res))
	engineRes := EngineResponse{Id: req.Id, B64Message: &b64res}

	outMessageBytes, _ := json.Marshal(engineRes)
	outMessage := string(outMessageBytes)

	//time measures
	logTime(start, "Response message marshaled", req.Id)

	err = qcOut.SendMsg(&outMessage, req.Id)

	//time measures
	logTime(start, "Total processing time", req.Id)

	if err == nil {
		qcIn.DeleteMessage(message.ReceiptHandle)
	}
}

func logTime(start time.Time, m string, id *string) {
	elapsed := time.Since(start)
	requestNumber := strings.Split(*id, "_")[0]
	log.Printf("Request '%s'. %s - %s", requestNumber, m, elapsed)
}
