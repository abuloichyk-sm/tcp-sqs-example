package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type HandleEngineResponseFunc func(res *EngineResponse)

type EngineClient struct {
	qcEngineIn           SqsQueueClient
	qcEngineOut          SqsQueueClient
	handleEngineResponse HandleEngineResponseFunc
	readQueueIntervalMs  int
}

func NewEngineClient(handleEngineResponse HandleEngineResponseFunc, readQueueIntervalMs int) *EngineClient {
	ec := &EngineClient{}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", "siliconmint"),
		Region:      aws.String("eu-central-1")},
	))

	ec.qcEngineIn = SqsQueueClient{}
	err := ec.qcEngineIn.Init(sess, aws.String("BuloichykEngineIn"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	ec.qcEngineOut = SqsQueueClient{}
	err = ec.qcEngineOut.Init(sess, aws.String("BuloichykEngineOut"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine out queue client created")

	ec.handleEngineResponse = handleEngineResponse
	ec.readQueueIntervalMs = readQueueIntervalMs
	return ec
}

func (ec *EngineClient) Run() {
	//read answers from engine
	go func() {
		t := time.NewTicker(time.Duration(ec.readQueueIntervalMs) * time.Millisecond)
		for {
			<-t.C
			out, _ := ec.qcEngineOut.ReceiveMessages()

			if len(out.Messages) == 0 {
				//debug output
				//log.Println("No messages")

				continue
			}

			log.Printf("Messages recived. Count %d", len(out.Messages))
			for _, mo := range out.Messages {
				go ec.processMessageFromEngine(mo)
			}
		}
	}()
}

func (ec *EngineClient) processMessageFromEngine(mo *sqs.Message) {
	log.Printf("Received from engine '%s'\n", *mo.Body)
	ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)

	engineRes := &EngineResponse{}
	err := json.Unmarshal([]byte(*mo.Body), engineRes)
	if err != nil {
		log.Printf("Error unmarshal '%s'\n", *mo.Body)
		ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
		return
	}

	ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
	ec.handleEngineResponse(engineRes)
}

func (ec *EngineClient) SendMessage(er *EngineRequest) {
	bytes, _ := json.Marshal(er)
	s := string(bytes)
	ec.qcEngineIn.SendMsg(&s, er.Id)
	log.Printf("Sent to engine '%s'\n", s)
}
