FROM scratch

COPY bin/mosquitto_exporter /mosquitto_exporter

EXPOSE 9324

ENTRYPOINT [ "/mosquitto_exporter" ]
