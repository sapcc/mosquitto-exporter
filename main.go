package main

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	appName   = "Mosquitto exporter"
	envPrefix = "MOSQUITTO_EXPORTER_"
)

var (
	mosquittoClient MQTTClient
	metrics         map[string]prometheus.Gauge
)

func main() {
	app := cli.NewApp()

	app.Name = appName
	app.Version = versionString()
	app.Authors = []cli.Author{
		{
			Name:  "Arturo Reuschenbach Puncernau",
			Email: "a.reuschenbach.puncernau@sap.com",
		},
	}
	app.Usage = "Mosquitto exporter"
	app.Action = runServer
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "Log level",
			EnvVar: envPrefix + "LOG_LEVEL",
			Value:  "info",
		},
		cli.StringSliceFlag{
			Name:   "endpoint,e",
			Usage:  "Endpoint url(s) for the mosquitto broker",
			EnvVar: envPrefix + "ENDPOINT",
			Value:  new(cli.StringSlice),
		},
		cli.StringFlag{
			Name:   "bind-address,b",
			Usage:  "Listen address for api server",
			Value:  "0.0.0.0:3000",
			EnvVar: envPrefix + "LISTEN",
		},
	}

	app.Before = func(c *cli.Context) error {
		// set log level
		lvl, err := log.ParseLevel(c.GlobalString("log-level"))
		if err != nil {
			log.Fatalf("Invalid log level: %s\n", c.GlobalString("log-level"))
			return err
		}
		log.SetLevel(lvl)
		return nil
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	// global transport instance
	mosquittoClient, err := mqttClient(c)
	if err != nil {
		log.Fatal(err)
	}
	defer mosquittoClient.Disconnect()

	// conect
	err = mosquittoClient.Connect()
	fatalfOnError(err, "Error connection to mosquitto broker.", err)

	metrics = map[string]prometheus.Gauge{}

	// save the environment
	go func() {
		brokerInfoChan, cancelBrokerInfoSubscription := mosquittoClient.Subscribe("$SYS/#")
		defer cancelBrokerInfoSubscription()

		for {
			select {
			case update := <-brokerInfoChan:
				processUpdate(string(update.Topic()), string(update.Payload()))
			}
		}
	}()

	// init the router and server
	router := newRouter()
	log.Infof("Listening on %q...", c.GlobalString("bind-address"))
	err = http.ListenAndServe(c.GlobalString("bind-address"), router)
	fatalfOnError(err, "Failed to bind on %s: ", c.GlobalString("bind-address"))
}

// $SYS/broker/bytes/received
func processUpdate(topic, payload string) {
	if metrics[topic] != nil {
		// update the first value
		value := parseValue(payload)
		metrics[topic].Set(value)
	} else {
		// ignore static metrics
		if topic != "$SYS/broker/timestamp" && topic != "$SYS/broker/version" {
			name := strings.Replace(topic, "$SYS/", "", 1)
			name = strings.Replace(name, "/", "_", -1)
			name = strings.Replace(name, " ", "_", -1)
			metrics[topic] = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: name,
				Help: topic,
			})
			// register the metric
			prometheus.MustRegister(metrics[topic])
			// add the first value
			value := parseValue(payload)
			metrics[topic].Set(value)
		}
	}
	log.Debugf("Got broker update with topic %s and data %s", topic, payload)
}

func parseValue(payload string) float64 {
	var validValue = regexp.MustCompile(`\d{1,}[.]\d{1,}|\d{1,}`)
	// get the first value of the string
	strArray := validValue.FindAllString(payload, 1)
	if len(strArray) > 0 {
		// parse to float
		value, err := strconv.ParseFloat(strArray[0], 64)
		if err == nil {
			return value
		}
	}
	return 0
}

func fatalfOnError(err error, msg string, args ...interface{}) {
	if err != nil {
		log.Fatalf(msg, args...)
	}
}
