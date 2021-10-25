package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/tls"

	"github.com/codegangsta/cli"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	appName = "Mosquitto exporter"
)

var (
	ignoreKeyMetrics = map[string]string{
		"$SYS/broker/timestamp":        "The timestamp at which this particular build of the broker was made. Static.",
		"$SYS/broker/version":          "The version of the broker. Static.",
		"$SYS/broker/clients/active":   "deprecated in favour of $SYS/broker/clients/connected",
		"$SYS/broker/clients/inactive": "deprecated in favour of $SYS/broker/clients/disconnected",
	}
	counterKeyMetrics = map[string]string{
		"$SYS/broker/bytes/received":            "The total number of bytes received since the broker started.",
		"$SYS/broker/bytes/sent":                "The total number of bytes sent since the broker started.",
		"$SYS/broker/messages/received":         "The total number of messages of any type received since the broker started.",
		"$SYS/broker/messages/sent":             "The total number of messages of any type sent since the broker started.",
		"$SYS/broker/publish/bytes/received":    "The total number of PUBLISH bytes received since the broker started.",
		"$SYS/broker/publish/bytes/sent":        "The total number of PUBLISH bytes sent since the broker started.",
		"$SYS/broker/publish/messages/received": "The total number of PUBLISH messages received since the broker started.",
		"$SYS/broker/publish/messages/sent":     "The total number of PUBLISH messages sent since the broker started.",
		"$SYS/broker/publish/messages/dropped":  "The total number of PUBLISH messages that have been dropped due to inflight/queuing limits.",
		"$SYS/broker/uptime":                    "The total number of seconds since the broker started.",
		"$SYS/broker/clients/maximum":           "The maximum number of clients connected simultaneously since the broker started",
		"$SYS/broker/clients/total":             "The total number of clients connected since the broker started.",
	}
	counterMetrics = map[string]*MosquittoCounter{}
	gaugeMetrics   = map[string]prometheus.Gauge{}
)

func main() {
	app := cli.NewApp()

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	mqtt.WARN = log.New(os.Stdout, "", 0)

	app.Name = appName
	app.Version = versionString()
	app.Authors = []cli.Author{
		{
			Name:  "Arturo Reuschenbach Puncernau",
			Email: "a.reuschenbach.puncernau@sap.com",
		},
		{
			Name:  "Fabian Ruff",
			Email: "fabian.ruff@sap.com",
		},
	}
	app.Usage = "Prometheus exporter for broker metrics"
	app.Action = runServer
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "endpoint,e",
			Usage:  "Endpoint for the Mosquitto message broker",
			EnvVar: "BROKER_ENDPOINT",
			Value:  "tcp://127.0.0.1:1883",
		},
		cli.StringFlag{
			Name:   "bind-address,b",
			Usage:  "Listen address for metrics HTTP endpoint",
			Value:  "0.0.0.0:9234",
			EnvVar: "BIND_ADDRESS",
		},
		cli.StringFlag{
			Name:   "user,u",
			Usage:  "Username for the Mosquitto message broker",
			Value:  "",
			EnvVar: "MQTT_USER",
		},
		cli.StringFlag{
			Name:   "pass,p",
			Usage:  "Password for the User on the Mosquitto message broker",
			Value:  "",
			EnvVar: "MQTT_PASS",
		},
		cli.StringFlag{
			Name:   "cert,c",
			Usage:  "Location of a TLS certificate .pem file for the Mosquitto message broker",
			Value:  "",
			EnvVar: "MQTT_CERT",
		},
		cli.StringFlag{
			Name:   "key,k",
			Usage:  "Location of a TLS private key .pem file for the Mosquitto message broker",
			Value:  "",
			EnvVar: "MQTT_KEY",
		},
		cli.StringFlag{
			Name:   "client-id,i",
			Usage:  "Client id to be used to connect to the Mosquitto message broker",
			Value:  "",
			EnvVar: "MQTT_CLIENT_ID",
		},
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	log.Printf("Starting mosquitto_broker %s", versionString())

	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(true)
	opts.AddBroker(c.String("endpoint"))

	if c.String("client-id") != "" {
		opts.SetClientID(c.String("client-id"))
	}

	// if you have a username you'll need a password with it
	if c.String("user") != "" {
		opts.SetUsername(c.String("user"))
		if c.String("pass") != "" {
			opts.SetPassword(c.String("pass"))
		}
	}
	// if you have a client certificate you want a key aswell
	if c.String("cert") != "" && c.String("key") != "" {
		keyPair, err := tls.LoadX509KeyPair(c.String("cert"), c.String("key"))
		if err != nil {
			log.Printf("Failed to load certificate/keypair: %s", err)
		}
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{keyPair},
			InsecureSkipVerify: true,
			ClientAuth:         tls.NoClientCert,
		}
		opts.SetTLSConfig(tlsConfig)
		if !strings.HasPrefix(c.String("endpoint"), "ssl://") &&
			!strings.HasPrefix(c.String("endpoint"), "tls://") {
			log.Println("Warning: To use TLS the endpoint URL will have to begin with 'ssl://' or 'tls://'")
		}
	} else if (c.String("cert") != "" && c.String("key") == "") ||
		(c.String("cert") == "" && c.String("key") != "") {
		log.Println("Warning: For TLS to work both certificate and private key are needed. Skipping TLS.")
	}

	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("Connected to %s", c.String("endpoint"))
		// subscribe on every (re)connect
		token := client.Subscribe("$SYS/#", 0, func(_ mqtt.Client, msg mqtt.Message) {
			processUpdate(msg.Topic(), string(msg.Payload()))
		})
		if !token.WaitTimeout(10 * time.Second) {
			log.Println("Error: Timeout subscribing to topic $SYS/#")
		}
		if err := token.Error(); err != nil {
			log.Printf("Failed to subscribe to topic $SYS/#: %s", err)
		}
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Error: Connection to %s lost: %s", c.String("endpoint"), err)
	}
	client := mqtt.NewClient(opts)

	// try to connect forever
	for {
		token := client.Connect()
		if token.WaitTimeout(5 * time.Second) {
			if token.Error() == nil {
				break
			}
			log.Printf("Error: Failed to connect to broker: %s", token.Error())
		} else {
			log.Printf("Timeout connecting to endpoint %s", c.String("endpoint"))
		}
		time.Sleep(5 * time.Second)
	}

	// init the router and server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", serveVersion)
	log.Printf("Listening on %s...", c.GlobalString("bind-address"))
	err := http.ListenAndServe(c.GlobalString("bind-address"), nil)
	fatalfOnError(err, "Failed to bind on %s: ", c.GlobalString("bind-address"))
}

// $SYS/broker/bytes/received
func processUpdate(topic, payload string) {
	//log.Printf("Got broker update with topic %s and data %s", topic, payload)
	if _, ok := ignoreKeyMetrics[topic]; !ok {
		if _, ok := counterKeyMetrics[topic]; ok {
			// log.Printf("Processing counter metric %s with data %s", topic, payload)
			processCounterMetric(topic, payload)
		} else {
			//log.Printf("Processing gauge metric %s with data %s", topic, payload)
			processGaugeMetric(topic, payload)
		}
	}
}

func processCounterMetric(topic, payload string) {
	if counterMetrics[topic] != nil {
		value := parseValue(payload)
		counterMetrics[topic].Set(value)
	} else {
		// create a mosquitto counter pointer
		mCounter := NewMosquittoCounter(prometheus.NewDesc(
			parseTopic(topic),
			topic,
			[]string{},
			prometheus.Labels{},
		))

		// save it
		counterMetrics[topic] = mCounter
		// register the metric
		prometheus.MustRegister(mCounter)
		// add the first value
		value := parseValue(payload)
		counterMetrics[topic].Set(value)
	}
}

func processGaugeMetric(topic, payload string) {
	if gaugeMetrics[topic] != nil {
		value := parseValue(payload)
		gaugeMetrics[topic].Set(value)
	} else {
		gaugeMetrics[topic] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: parseTopic(topic),
			Help: topic,
		})
		// register the metric
		prometheus.MustRegister(gaugeMetrics[topic])
		// add the first value
		value := parseValue(payload)
		gaugeMetrics[topic].Set(value)
	}
}

func parseTopic(topic string) string {
	name := strings.Replace(topic, "$SYS/", "", 1)
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, " ", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return name
}

func parseValue(payload string) float64 {
	// fmt.Printf("Payload %s \n", payload)
	var validValue = regexp.MustCompile(`-?\d{1,}[.]\d{1,}|\d{1,}`)
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
