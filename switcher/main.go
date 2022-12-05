package main

func main() {
	sw := Switcher{}

	ec := &EngineClient{}
	ec.Init(sw.HandleEngineResponse)
	ec.Run()

	ts := &TcpServer{}
	ts.Init(sw.HandleTcpRequest)

	sw.Init(ts, ec)

	ts.Run()
}
