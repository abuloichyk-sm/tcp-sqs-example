package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			//conn, _ := net.Dial("tcp", "127.0.0.1:8081")
			conn, _ := net.Dial("tcp", "3.64.255.146:8082")

			//request just number
			req := fmt.Sprintf("%d_%v", i, time.Now().Unix())

			//or request with time to pass duplication check in queue
			//req := fmt.Sprintf("%d %d", i, time.Now().Unix())

			start := time.Now()
			// Отправляем в socket
			fmt.Fprint(conn, req)
			fmt.Printf("sent %s\n", req)

			// Прослушиваем ответ
			res, _ := bufio.NewReader(conn).ReadString('\n')

			elapsed := time.Since(start)
			log.Printf("Process %d took %s", i, elapsed)
			conn.Close()

			fmt.Printf("Request: '%s', response: '%s'\n", req, res)
		}()
	}
	wg.Wait()
}

func ReadFromStdIn() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Text to send: ")
	text, _ := reader.ReadString('\n')
	text = text[:len(text)-2] //remove new line 2 char in windows
	return text
}
