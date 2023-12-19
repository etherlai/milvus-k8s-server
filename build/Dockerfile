FROM golang:1.20 as builder
WORKDIR /go/src/milvus-k8s-server
COPY . .
RUN go build -ldflags "-s -w -X main.version=v0.1" -o build/server cmd/main.go

FROM golang:1.20

RUN apt-get update && apt-get install -y tzdata && ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
WORKDIR /milvus
COPY --from=builder  /go/src/milvus-k8s-server/build/server server
ENTRYPOINT ["./server"]
