package queueclient

type EngineRequest struct {
	Id         *string
	B64Message *string
}

type EngineResponse struct {
	Id         *string
	B64Message *string
}
