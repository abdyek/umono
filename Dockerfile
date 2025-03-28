FROM golang:1.23

WORKDIR /usr/src/app

RUN touch umono.db

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./

CMD ["app"]
