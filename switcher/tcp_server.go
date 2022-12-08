package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

type HandleTcpRequestFunc func(m *string, conn *net.Conn)

type TcpServer struct {
	handleReqFunc HandleTcpRequestFunc
	port          int
}

func NewTcpServer(handleReqFunc HandleTcpRequestFunc, port int) *TcpServer {
	ts := &TcpServer{}

	ts.handleReqFunc = handleReqFunc
	ts.port = port
	return ts
}

func (ts *TcpServer) Run() {
	fmt.Println("Launching server...")

	listen, err := net.Listen("tcp", ":"+fmt.Sprint(ts.port))
	if err != nil {
		log.Printf("error %v", err)
	}
	defer listen.Close()

	fmt.Println("Server launched on " + fmt.Sprint(ts.port))

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
	req := readMessage(conn)

	//log
	log.Printf("Received from tcp '%s'\n", *req)

	ts.handleReqFunc(req, &conn)
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

func (ts *TcpServer) WriteResponse(res *string, conn *net.Conn) {
	(*conn).Write([]byte(*res))
	(*conn).Close()
	log.Printf("Sent to tcp '%s'\n", *res)
}
