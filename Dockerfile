FROM golang:1.20.0-bullseye
COPY *.go /app/
COPY go.* /app/
WORKDIR /app
RUN go build -o exe
ENTRYPOINT ["/app/exe"]
