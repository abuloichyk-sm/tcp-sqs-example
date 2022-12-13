package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type RunResult struct {
	Index            int
	AvgMillisec      int
	MaxElapsed       time.Duration
	MaxRequest       *RequestRun
	TimeStart        time.Time
	MessagesInSecond int
	IntervalSeconds  int
	Executed         bool
}

type RequestRun struct {
	Elapsed  time.Duration
	Request  string
	Response string
}

var RunResults []*RunResult = make([]*RunResult, 0)
var mu sync.Mutex

func main() {
	http.HandleFunc("/allRuns", allRunsHandler)
	http.HandleFunc("/newRun", newRunHandler)
	http.HandleFunc("/", rootHandler)

	go Runner()

	http.ListenAndServe(":8085", nil)
}

func Runner() {
	lastRunnedIndex := -1
	for {
		if len(RunResults) > lastRunnedIndex+1 {
			indexToRun := lastRunnedIndex + 1
			RunResults[indexToRun].ExecuteRequests("3.64.255.146:8082")

			lastRunnedIndex = indexToRun
		} else {
			time.Sleep(time.Second)
		}
	}
}

func allRunsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<body>")
	defer writeFooter(w)

	fmt.Fprintf(w, "All runs:</br></br>")
	for i := 0; i < len(RunResults); i++ {
		writeRun(w, RunResults[i])
		fmt.Fprintf(w, "-------------------</br>")
	}
}

func newRunHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New run " + r.URL.String())
	fmt.Fprintf(w, "<body>")
	defer writeFooter(w)

	messagesInSecondParam := r.URL.Query().Get("messagesInSecond")
	messagesInSecond, err := strconv.Atoi(messagesInSecondParam)
	if err != nil {
		messagesInSecond = 10
	}

	intervalSecondsParam := r.URL.Query().Get("intervalSeconds")
	intervalSeconds, err := strconv.Atoi(intervalSecondsParam)
	if err != nil {
		messagesInSecond = 1
	}

	mu.Lock()
	index := len(RunResults)
	RunResults = append(RunResults, &RunResult{
		Index:            index,
		MessagesInSecond: messagesInSecond,
		IntervalSeconds:  intervalSeconds,
	})
	mu.Unlock()

	// go func() {
	// 	RunResults[index] = executeRequests("3.64.255.146:8082", messagesInSecond, intervalSeconds)
	// }()

	fmt.Fprintf(w, "Hello! Run queued. Results will be located <a href=\"/%d\">here</a>", index)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Root " + r.URL.String())
	fmt.Fprintf(w, "<body>")
	defer writeFooter(w)

	urlPart := r.URL.String()[1:]
	index, err := strconv.Atoi(urlPart)
	if err != nil {
		fmt.Fprintf(w, "Not valid run index %s", urlPart)
		return
	}
	if index >= len(RunResults) {
		fmt.Fprintf(w, "There are only %d runs", len(RunResults))
		return
	}

	writeRun(w, RunResults[index])
}

func writeRun(w http.ResponseWriter, res *RunResult) {
	fmt.Fprintf(w, "Run %d results:</br>", res.Index)
	fmt.Fprintf(w, "TimeStarted: %s</br>", res.TimeStart.Format(time.UnixDate))
	fmt.Fprintf(w, "Executed: %t</br>", res.Executed)
	fmt.Fprintf(w, "Messages in second: %d</br>", res.MessagesInSecond)
	fmt.Fprintf(w, "Measured interval: %ds</br>", res.IntervalSeconds)
	fmt.Fprintf(w, "Avg request time: %dms</br>", res.AvgMillisec)
	fmt.Fprintf(w, "Max request time: %s</br>", res.MaxElapsed)
	if res.MaxRequest != nil {
		fmt.Fprintf(w, "Max request body: %s</br>", res.MaxRequest.Request)
	}
}

func writeFooter(w http.ResponseWriter) {
	fmt.Fprintf(w, "</br>")
	fmt.Fprintf(w, "</br><a href=\"/newRun?messagesInSecond=10&intervalSeconds=1\">New run</a>")
	fmt.Fprintf(w, "</br><a href=\"/allRuns\">All runs result</a>")
	fmt.Fprintf(w, "</body>")
}

func (res *RunResult) ExecuteRequests(tcpAddress string) {
	res.TimeStart = time.Now()
	wg := sync.WaitGroup{}

	bulkSize := int(float64(res.MessagesInSecond) / float64(10))
	iterationsCount := res.IntervalSeconds * 10
	requestsCount := bulkSize * iterationsCount

	runs := make([]*RequestRun, requestsCount)

	var messageCounter int32 = -1 // so first be 0
	var mu sync.Mutex

	for i := 0; i < iterationsCount; i++ {
		time.Sleep(100 * time.Millisecond)

		for j := 0; j < bulkSize; j++ {
			index := atomic.AddInt32(&messageCounter, 1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				//conn, _ := net.Dial("tcp", "127.0.0.1:80")
				conn, _ := net.Dial("tcp", tcpAddress)

				// request with time to pass duplication check in queue
				//req := fmt.Sprintf("%d %d", i, time.Now().Unix())

				//request with transaction
				req := fmt.Sprintf("%d_%s", index, "019025D0B18001002   1011083157382     20221107VD0ENT    AM01C412345678C51CATESTYCBTESTERSONCM123 FIRST STCNSALT LAKE CITYCOUTCP84103CQ5555555555C7014X1AM04C211111111C1GIDC61AM07EM1D24009584E103D765862067605E730000D30D530D61D80DE20221027DF0DJ3ET30000DT1EB3000028EAU71AM03EZ01DB1649398694DRWITHERSPM80148316002JINEZ2K1208 E 3300 S2MSALT LAKE CTY2NUT2P84106AM11D9208EDC100{DQ340CDU308EDN01")

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
				runs[index] = &RequestRun{
					Elapsed:  elapsed,
					Request:  req,
					Response: res,
				}
				mu.Unlock()
			}()
		}
	}
	wg.Wait()

	res.CalculateMetrics(runs)
	res.Executed = true
}

func (runRes *RunResult) CalculateMetrics(runs []*RequestRun) {
	maxIndex := 0
	totalDuration := time.Duration(0)
	for i := 0; i < len(runs); i++ {
		totalDuration += runs[i].Elapsed
		if runs[maxIndex].Elapsed < runs[i].Elapsed {
			maxIndex = i
		}
	}

	avgDurationMilliseconds := int(float64(totalDuration.Milliseconds()) / float64(len(runs)))

	runRes.AvgMillisec = avgDurationMilliseconds
	runRes.MaxElapsed = runs[maxIndex].Elapsed
	runRes.MaxRequest = runs[maxIndex]
}
