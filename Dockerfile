FROM golang:1.17-buster AS build

WORKDIR /go/src/app

## Download modules and store, this optimizes use of Docker image cache
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN make build

FROM scratch
LABEL source_repository="https://github.com/sapcc/mosquitto-exporter"

COPY --from=build /go/src/app/bin/mosquitto_exporter /mosquitto_exporter

EXPOSE 9234

ENTRYPOINT [ "/mosquitto_exporter" ]
