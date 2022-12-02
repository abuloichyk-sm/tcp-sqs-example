package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SqsQueueClient struct {
	Session           *session.Session
	QueueUrl          *string
	VisibilityTimeout *int64 //seconds
	svc               *sqs.SQS
}

func (qc *SqsQueueClient) Init(sess *session.Session, queueName *string, visibilityTimeout *int64) error {
	qc.Session = sess
	qc.svc = sqs.New(sess)

	queueUrl, err := qc.getQueueURL(sess, queueName)
	qc.QueueUrl = queueUrl
	qc.VisibilityTimeout = visibilityTimeout
	if err != nil {
		return err
	}
	return nil
}

func (qc *SqsQueueClient) getQueueURL(sess *session.Session, queue *string) (*string, error) {
	urlResult, err := qc.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queue,
	})

	if err != nil {
		return nil, err
	}

	return urlResult.QueueUrl, nil
}

func (qc *SqsQueueClient) SendMsg(body *string) error {
	_, err := qc.svc.SendMessage(&sqs.SendMessageInput{
		MessageBody:    body,
		MessageGroupId: aws.String("EngineQueue"), //Messages in one group will be ordered as FIFO. Only one group this case
		QueueUrl:       qc.QueueUrl,
	})
	if err != nil {
		log.Printf("Error on message sent body'%s' %d err %s", *body, len(*body), err)
		return err
	}

	return nil
}

func (qc *SqsQueueClient) ReceiveMessages() (*sqs.ReceiveMessageOutput, error) {
	msgResult, err := qc.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            qc.QueueUrl,
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   qc.VisibilityTimeout,
	})

	if err != nil {
		return nil, err
	}

	return msgResult, nil
}

func (qc *SqsQueueClient) DeleteMessage(messageHandle *string) error {
	_, err := qc.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      qc.QueueUrl,
		ReceiptHandle: messageHandle,
	})
	if err != nil {
		return err
	}

	return nil
}
