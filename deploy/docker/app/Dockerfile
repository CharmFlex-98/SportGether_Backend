ARG GOLANG_IMAGE="golang:1.21.3"
ARG BASE_IMAGE="gcr.io/distroless/base-debian12:nonroot"

FROM $GOLANG_IMAGE as builder
WORKDIR /sportgether
COPY ../../../go.mod ./
RUN go mod download

COPY . ./
WORKDIR ./cmd/api
RUN env GOOS=linux GOARCH=arm64 go build -o ./sportgether_binary

FROM $BASE_IMAGE
COPY --from=builder /sportgether/cmd/api/sportgether_binary ./

CMD ["./sportgether_binary"]
