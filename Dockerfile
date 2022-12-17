# Get base dependency for golang
FROM golang:1.19.3-alpine

# maintainer info
LABEL maintainer = "Abhijith A <abhijithak683@gmail.com>"

WORKDIR /app

COPY ./msghub-server/go.mod ./

RUN go mod tidy

COPY . .

RUN cd msghub-server && go build -o main

EXPOSE 9000

