Mosquitto Exporter
=============================
Prometheus exporter for the [Mosquitto MQTT message broker](https://mosquitto.org/).

There is a docker image available:
```
docker run \
  -p 9234:9234 sapcc/mosquitto-exporter \
  --endpoint tcp://mosquitto:1883
```

Usage:
======

```
NAME:
   Mosquitto exporter - Prometheus exporter for broker metrics

USAGE:
   mosquitto_exporter [global options] command [command options] [arguments...]

VERSION:
   0.6.0 (0ac92b5), go1.14.2

AUTHORS:
   Arturo Reuschenbach Puncernau <a.reuschenbach.puncernau@sap.com>
   Fabian Ruff <fabian.ruff@sap.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --endpoint value, -e value      Endpoint for the Mosquitto message broker (default: "tcp://127.0.0.1:1883") [$BROKER_ENDPOINT]
   --bind-address value, -b value  Listen address for metrics HTTP endpoint (default: "0.0.0.0:9234") [$BIND_ADDRESS]
   --user value, -u value          Username for the Mosquitto message broker [$MQTT_USER]
   --pass value, -p value          Password for the User on the Mosquitto message broker [$MQTT_PASS]
   --cert value, -c value          Location of a TLS certificate .pem file for the Mosquitto message broker [$MQTT_CERT]
   --key value, -k value           Location of a TLS private key .pem file for the Mosquitto message broker [$MQTT_KEY]
   --help, -h                      show help
   --version, -v                   print the version
```
