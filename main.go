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
	mosquittoClient  MQTTClient
	ignoreKeyMetrics = map[string]string{
		"$SYS/broker/timestamp": "The timestamp at which this particular build of the broker was made. Static.",
		"$SYS/broker/version":   "The version of the broker. Static.",
	}
	counterKeyMetrics = map[string]string{
		"$SYS/broker/bytes/received":            "The total number of bytes received since the broker started.",
		"$SYS/broker/bytes/sent":                "The total number of bytes sent since the broker started.",
		"$SYS/broker/messages/received":         "The total number of messages of any type received since the broker started.",
		"$SYS/broker/messages/sent":             "The total number of messages of any type sent since the broker started.",
		"$SYS/broker/publish/messages/received": "The total number of PUBLISH messages received since the broker started.",
		"$SYS/broker/publish/messages/sent":     "The total number of PUBLISH messages sent since the broker started.",
	}
	counterMetrics = map[string]prometheus.Counter{}
	gougeMetrics   = map[string]prometheus.Gauge{}
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
	log.Debugf("Got broker update with topic %s and data %s", topic, payload)
	if _, ok := ignoreKeyMetrics[topic]; !ok {
		if _, ok := counterKeyMetrics[topic]; ok {
			log.Debugf("Processing counter metric %s with data %s", topic, payload)
			processCounterMetric(topic, payload)
		} else {
			log.Debugf("Processing gauge metric %s with data %s", topic, payload)
			processGaugeMetric(topic, payload)
		}
	}
}

func processCounterMetric(topic, payload string) {
	if counterMetrics[topic] != nil {
		value := parseValue(payload)
		counterMetrics[topic].Add(value)
	} else {
		counterMetrics[topic] = prometheus.NewCounter(prometheus.CounterOpts{
			Name: parseTopic(topic),
			Help: topic,
		})
		// register the metric
		prometheus.MustRegister(counterMetrics[topic])
		// add the first value
		value := parseValue(payload)
		counterMetrics[topic].Add(value)
	}
}

func processGaugeMetric(topic, payload string) {
	if gougeMetrics[topic] != nil {
		value := parseValue(payload)
		gougeMetrics[topic].Set(value)
	} else {
		gougeMetrics[topic] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: parseTopic(topic),
			Help: topic,
		})
		// register the metric
		prometheus.MustRegister(gougeMetrics[topic])
		// add the first value
		value := parseValue(payload)
		gougeMetrics[topic].Set(value)
	}
}

func parseTopic(topic string) string {
	name := strings.Replace(topic, "$SYS/", "", 1)
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, " ", "_", -1)
	return name
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
