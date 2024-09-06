BINARY_NAME=vdex

build:
 GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin main.go
 GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
 GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows main.go
 go build -o ${BINARY_NAME} main.go

run: build
 ./${BINARY_NAME}

clean:
 go clean
 rm ${BINARY_NAME}-darwin
 rm ${BINARY_NAME}-linux
 rm ${BINARY_NAME}-windows
 rm ./${BINARY_NAME}

test:
 go test ./...

test_coverage:
 go test ./... -coverprofile=coverage.out