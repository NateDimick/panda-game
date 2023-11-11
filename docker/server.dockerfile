FROM golang:1.21.3-alpine

COPY backend/ /panda-game-workdir/

WORKDIR /panda-game-workdir

RUN go version

RUN go mod download

RUN CGO_ENABLED=0 go build -o panda-game-server cmd/panda-game/main.go

FROM alpine

COPY --from=0 /panda-game-workdir/panda-game-server /usr/local/bin/

CMD [ "panda-game-server" ]