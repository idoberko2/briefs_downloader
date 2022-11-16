BINARY_NAME=dnldr

build:
	go build -o ${BINARY_NAME} ./...

buildpi:
	GOOS=linux GOARCH=arm GOARM=5 go build -o ${BINARY_NAME} ./...
