package main

import (
	"encoding/base64"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	queueclient "github.com/abuloichyk-sm/tcp-sqs-example/internal/queueclient"
)

type SendMessageToQueueFunc func(req *queueclient.EngineRequest)
type SendToTcpConnFunc func(res *string, conn *net.Conn)

type SwitcherRequest struct {
	Id           *string
	Message      *string
	Conn         *net.Conn
	ProcessStart time.Time
}

type Switcher struct {
	Requests sync.Map
	ts       *TcpServer
	ec       *EngineClient
}

func NewSwitcher() *Switcher {
	sw := &Switcher{}

	sw.ec = NewEngineClient(sw.HandleEngineResponse)
	sw.ts = NewTcpServer(sw.HandleTcpRequest)

	return sw
}

func (sw *Switcher) Run() {
	sw.ec.Run()
	sw.ts.Run()
}

func (sw *Switcher) HandleTcpRequest(m *string, conn *net.Conn) {
	id := *m + "_" + uuid.New().String()
	sr := SwitcherRequest{
		Id:           &id,
		Message:      m,
		Conn:         conn,
		ProcessStart: time.Now(),
	}

	sw.Requests.Store(*sr.Id, sr)
	log.Printf("Stored in map with id '%s', message '%s'", *sr.Id, *m)

	b64Message := base64.StdEncoding.EncodeToString([]byte(*m))
	er := &queueclient.EngineRequest{
		Id:         sr.Id,
		B64Message: &b64Message,
	}
	sw.ec.SendMessage(er)
}

func (sw *Switcher) HandleEngineResponse(res *queueclient.EngineResponse) {
	decodedBytes, err := base64.StdEncoding.DecodeString(*res.B64Message)
	if err != nil {
		log.Printf("Base 64 corrupted for message id '%s'. Base 64 value '%s', decoded '%s', error '%v'",
			*res.Id, *res.B64Message, decodedBytes, err)
		return
	}
	message := string(decodedBytes)

	srAny, ok := sw.Requests.Load(*res.Id)
	if !ok {
		log.Printf("Not found in map conn for Id '%s'. Message '%s'", *res.Id, message)
		return
	}

	sr, _ := srAny.(SwitcherRequest)

	//measures
	elapsed := time.Since(sr.ProcessStart)
	requestNumber := strings.Split(*sr.Id, "_")[0]
	log.Printf("Request '%s' total processing time - %s. Id '%s'", requestNumber, elapsed, *sr.Id)

	sw.ts.WriteResponse(&message, sr.Conn)
}
