FROM golang:alpine

COPY backend/ /panda-game-workdir/

WORKDIR /panda-game-workdir

RUN go version

RUN go mod download

RUN CGO_ENABLED=0 go build -o panda-nats-init cmd/nats-init/main.go

FROM alpine

COPY --from=0 /panda-game-workdir/panda-nats-init /usr/local/bin/

CMD [ "panda-nats-init" ]