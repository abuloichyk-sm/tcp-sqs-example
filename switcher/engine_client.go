package main

import (
	"encoding/json"
	"log"
	"time"

	queueclient "github.com/abuloichyk-sm/tcp-sqs-example/internal/queueclient"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type HandleEngineResponseFunc func(res *queueclient.EngineResponse)

type EngineClient struct {
	qcEngineIn           queueclient.SqsQueueClient
	qcEngineOut          queueclient.SqsQueueClient
	handleEngineResponse HandleEngineResponseFunc
}

func NewEngineClient(handleEngineResponse HandleEngineResponseFunc) *EngineClient {
	ec := &EngineClient{}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", "siliconmint"),
		Region:      aws.String("eu-central-1")},
	))

	ec.qcEngineIn = queueclient.SqsQueueClient{}
	err := ec.qcEngineIn.Init(sess, aws.String("BuloichykEngineIn"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	ec.qcEngineOut = queueclient.SqsQueueClient{}
	err = ec.qcEngineOut.Init(sess, aws.String("BuloichykEngineOut"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine out queue client created")

	ec.handleEngineResponse = handleEngineResponse
	return ec
}

func (ec *EngineClient) Run() {
	//read answers from engine
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
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

	engineRes := &queueclient.EngineResponse{}
	err := json.Unmarshal([]byte(*mo.Body), engineRes)
	if err != nil {
		log.Printf("Error unmarshal '%s'\n", *mo.Body)
		ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
		return
	}

	ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
	ec.handleEngineResponse(engineRes)
}

func (ec *EngineClient) SendMessage(er *queueclient.EngineRequest) {
	bytes, _ := json.Marshal(er)
	s := string(bytes)
	ec.qcEngineIn.SendMsg(&s, er.Id)
	log.Printf("Sent to engine '%s'\n", s)
}
