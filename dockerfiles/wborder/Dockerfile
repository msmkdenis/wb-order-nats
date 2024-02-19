FROM golang:1.21.3-alpine3.17

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o ./wborder ./cmd/wborder/main.go

# command to be used to execute when the image is used to start a container
CMD [ "./wborder" ]