package main

import (
	"sync"
)

func main() {
	chans := &sync.Map{}
	toProcess := make(chan (string), 1)

	ec := &EngineClient{}
	ec.Init(chans, toProcess)
	ec.Run()

	ts := &TcpServer{}
	ts.Init(chans, toProcess)
	ts.Run()
}
