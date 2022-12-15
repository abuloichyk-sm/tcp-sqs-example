aws lambda invoke --function-name main out.json

aws lambda invoke --region eu-central-1 --function-name PahrmaCoPocBuloichykLambda out.json

local run (install sam)
sam build
sam local invoke PahrmaCoPocBuloichykLambda -e event.json

build and deploy:
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o ./build/main main.go engine_models.go sqs_queue_client.go
~\Go\Bin\build-lambda-zip.exe -o ./build/main.zip ./build/main