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
}

func (ts *TcpServer) Init(handleReqFunc HandleTcpRequestFunc) {
	ts.handleReqFunc = handleReqFunc
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
