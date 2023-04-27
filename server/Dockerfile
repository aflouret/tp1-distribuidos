FROM golang:1.19 AS builder

WORKDIR server/

COPY go.mod ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server

# Deploy the application binary into a lean image
FROM busybox:latest
COPY --from=builder /bin/server /server
ENTRYPOINT ["/bin/sh"]