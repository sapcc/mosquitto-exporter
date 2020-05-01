FROM golang:1.14.2-buster AS build

WORKDIR /go/src/app

COPY . .

RUN make build

FROM scratch

COPY --from=build /go/src/app/bin/mosquitto_exporter /mosquitto_exporter

EXPOSE 9234

ENTRYPOINT [ "/mosquitto_exporter" ]
