package main

import (
	"encoding/base64"
	"log"
	"net"
	"sync"

	queueclient "github.com/abuloichyk-sm/tcp-sqs-example/internal/queueclient"

	"github.com/google/uuid"
)

type SendMessageToQueueFunc func(req *queueclient.EngineRequest)
type SendToTcpConnFunc func(res *string, conn *net.Conn)

type SwitcherRequest struct {
	Id      *string
	Message *string
	Conn    *net.Conn
}

type Switcher struct {
	Requests          sync.Map
	sendMessageFunc   SendMessageToQueueFunc
	sendToTcpConnFunc SendToTcpConnFunc
	ts                *TcpServer
}

func (sw *Switcher) Init(sendMessageFunc SendMessageToQueueFunc, sendToTcpConnFunc SendToTcpConnFunc) {
	sw.sendMessageFunc = sendMessageFunc
	sw.sendToTcpConnFunc = sendToTcpConnFunc
}

func (sw *Switcher) HandleTcpRequest(m *string, conn *net.Conn) {
	id := uuid.New().String()
	sr := SwitcherRequest{
		Id:      &id,
		Message: m,
		Conn:    conn,
	}

	sw.Requests.Store(*sr.Id, sr)
	log.Printf("Stored for id %s", *sr.Id)

	b64Message := base64.StdEncoding.EncodeToString([]byte(*m))
	er := &queueclient.EngineRequest{
		Id:         sr.Id,
		B64Message: &b64Message,
	}
	sw.sendMessageFunc(er)
}

func (sw *Switcher) HandleEngineResponse(res *queueclient.EngineResponse) {
	resBytes, err := base64.StdEncoding.DecodeString(*res.B64Message)
	if err != nil {
		log.Printf("Base 64 corrupted for message id '%s'. Base 64 value '%s', resBytes '%s',err  '%v'", *res.Id, *res.B64Message, resBytes, err)
		return
	}
	message := string(resBytes)

	srAny, ok := sw.Requests.Load(*res.Id)
	if !ok {
		log.Printf("Not found conn for Id '%s'. Message '%s'", *res.Id, message)
		return
	}

	sr, _ := srAny.(SwitcherRequest)

	sw.sendToTcpConnFunc(&message, sr.Conn)
}
