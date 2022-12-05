package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	queueclient "github.com/abuloichyk-sm/tcp-sqs-example/internal/queueclient"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type EngineClient struct {
	qcEngineIn  queueclient.SqsQueueClient
	qcEngineOut queueclient.SqsQueueClient
	chans       *sync.Map
	toProcess   chan (string)
}

func (ec *EngineClient) Init(chans *sync.Map, toProcess chan (string)) {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", ""),
		Region:      aws.String("us-east-1")},
	))

	ec.qcEngineIn = queueclient.SqsQueueClient{}
	err := ec.qcEngineIn.Init(sess, aws.String("AlexTestQueueEngineIn.fifo"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine in queue client created")

	ec.qcEngineOut = queueclient.SqsQueueClient{}
	err = ec.qcEngineOut.Init(sess, aws.String("AlexTestQueueEngineOut.fifo"), aws.Int64(1))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Engine out queue client created")

	ec.chans = chans
	ec.toProcess = toProcess
}

func (ec *EngineClient) Run() {
	//send messages to engine
	go func() {
		for {
			m := <-ec.toProcess
			deduplicaionId := fmt.Sprintf("%s_%d", m, time.Now().Unix())
			ec.qcEngineIn.SendMsg(&m, &deduplicaionId)
			log.Printf("Sent to engine '%s'\n", m)
		}
	}()

	//read answers from engine
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			<-t.C
			out, _ := ec.qcEngineOut.ReceiveMessages()
			if len(out.Messages) == 0 {
				log.Println("No messages")
				continue
			}
			log.Printf("Messages count %d", len(out.Messages))
			for _, mo := range out.Messages {
				go ec.processMessageFromEngine(mo)
			}
		}
	}()
}

func (ec *EngineClient) processMessageFromEngine(mo *sqs.Message) {
	log.Printf("Received from engine '%s'\n", *mo.Body)

	engineRes := &queueclient.EngineResponse{}
	err := json.Unmarshal([]byte(*mo.Body), engineRes)
	if err != nil {
		log.Printf("Error unmarshal '%s'\n", *mo.Body)
		ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
		return
	}

	chanForRes, ok := ec.chans.Load(engineRes.Id)
	if !ok {
		log.Printf("Chan for key %s not found\n", engineRes.Id)
		ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
		return
	}
	c := chanForRes.(chan (string))
	c <- engineRes.Message

	ec.qcEngineOut.DeleteMessage(mo.ReceiptHandle)
}
