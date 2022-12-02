package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _ := net.Dial("tcp", "127.0.0.1:8081")

			req := fmt.Sprintf("%d %d", i, time.Now().Unix())
			// Отправляем в socket
			fmt.Fprint(conn, req)
			fmt.Printf("sent %d\n", i)

			// Прослушиваем ответ
			res, _ := bufio.NewReader(conn).ReadString('\n')
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
