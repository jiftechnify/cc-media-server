FROM golang:latest AS coreBuilder
WORKDIR /work

COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
RUN go build -o mediaserver

FROM golang:latest

RUN apt-get update && apt-get install -y libwebp-dev \
    && rm -rf /var/lib/apt/lists/*
COPY --from=coreBuilder /work/mediaserver /usr/local/bin

CMD ["mediaserver"]
