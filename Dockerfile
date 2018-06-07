FROM scratch

COPY bin/mosquitto_exporter /mosquitto_exporter

EXPOSE 9234

ENTRYPOINT [ "/mosquitto_exporter" ]
