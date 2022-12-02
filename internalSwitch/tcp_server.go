package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type TcpServer struct {
	chans     *sync.Map
	toProcess chan (string)
}

func (ts *TcpServer) Init(chans *sync.Map, toProcess chan (string)) {
	ts.chans = chans
	ts.toProcess = toProcess
}

func (ts *TcpServer) Run() {
	fmt.Println("Launching server...")

	port := 8081
	listen, _ := net.Listen("tcp", ":"+fmt.Sprint(port))
	defer listen.Close()

	fmt.Println("Server launched on " + fmt.Sprint(port))

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go ts.handleIncomingRequest(conn)
	}
}

func (ts *TcpServer) handleIncomingRequest(conn net.Conn) {
	defer conn.Close()
	req := readMessage(conn)

	//log
	log.Printf("Received from tcp '%s'\n", *req)

	//mock response
	//res := fmt.Sprintf("%s processed", *req)

	//send to engine
	var c = make(chan (string), 1)
	ts.chans.Store(*req, c)
	ts.toProcess <- *req

	//wait for response from engine
	res := <-c
	ts.chans.Delete(*req)
	close(c)

	writeResponse(conn, &res)
	log.Printf("Sent to tcp '%s'\n", res)
}

func readMessage(conn net.Conn) *string {
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	res := string(buffer[:count])
	return &res
}

func writeResponse(conn net.Conn, message *string) {
	conn.Write([]byte(*message))
}
