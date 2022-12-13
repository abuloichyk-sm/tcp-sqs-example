package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// time := fmt.Sprintf("%s", time.Now().Format(time.StampMilli))
	// fmt.Println(time)
	wg := sync.WaitGroup{}

	messagesInSecond := 40
	intervalSeconds := 60

	bulkSize := int(float64(messagesInSecond) / float64(10))
	iterationsCount := intervalSeconds * 10

	var messageCounter int32 = 0

	var maxElapsed = time.Duration(0)
	var maxReq string
	var mu sync.Mutex

	for i := 0; i < iterationsCount; i++ {
		time.Sleep(100 * time.Millisecond)

		for j := 0; j < bulkSize; j++ {
			index := atomic.AddInt32(&messageCounter, 1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				//conn, _ := net.Dial("tcp", "127.0.0.1:80")
				conn, _ := net.Dial("tcp", "3.64.255.146:8082")

				//request just number
				req := fmt.Sprintf("%d_%v", index, time.Now().Unix())

				//or request with time to pass duplication check in queue
				//req := fmt.Sprintf("%d %d", i, time.Now().Unix())

				start := time.Now()
				// Отправляем в socket
				fmt.Fprint(conn, req)
				fmt.Printf("sent %s\n", req)

				// Прослушиваем ответ
				res, _ := bufio.NewReader(conn).ReadString('\n')

				elapsed := time.Since(start)
				log.Printf("Process %d took %s", index, elapsed)
				conn.Close()

				fmt.Printf("Request: '%s', response: '%s' start time: %s\n", req, res, start.Format(time.StampMilli))

				mu.Lock()
				if elapsed > maxElapsed {
					maxElapsed = elapsed
					maxReq = req
				}
				mu.Unlock()
			}()
		}
	}
	wg.Wait()
	fmt.Printf("Max time: %s. Request: '%s'\n", maxElapsed, maxReq)
}

func ReadFromStdIn() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Text to send: ")
	text, _ := reader.ReadString('\n')
	text = text[:len(text)-2] //remove new line 2 char in windows
	return text
}
