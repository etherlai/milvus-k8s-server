FROM golang:1.14 as builder
WORKDIR /go/src/milvus-k8s-server
COPY . .
RUN go build -ldflags "-s -w -X main.version=v0.1" -o build/server cmd/main.go

FROM ubuntu:18.04

RUN apt-get update && apt-get install -y tzdata && ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
WORKDIR /milvus
COPY --from=builder  /go/src/dce/milvus-k8s-server/server server
ENTRYPOINT ["./server"]
