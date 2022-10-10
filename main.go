package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/pvainio/scd30"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/kelseyhightower/envconfig"

	mqttClient "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	MqttUrl          string `envconfig:"mqtt_url" required:"true"`
	MqttUser         string `envconfig:"mqtt_user"`
	MqttPwd          string `envconfig:"mqtt_password"`
	MqttClientId     string `envconfig:"mqtt_client_id" default:"scd30"`
	Interval         uint16 `envconfig:"interval" default:"50"`
	AutoCalibration  uint16 `envconfig:"autocal" default:"1"`
	ForceCalibration uint16 `envconfig:"forcecal"`
	TempOffset       uint16 `envconfig:"temp_offset" default:"150"`
	Id               string `envconfig:"id" default:"scd30"`
	Name             string `envconfig:"name" default:"SCD30"`
	Debug            bool   `envconfig:"debug" default:"false"`
}

type measurement struct {
	id     string
	time   time.Time
	value  float32
	format string
	min    float32
	max    float32
}

var (
	config Config

	logInfo  *log.Logger
	logDebug *log.Logger

	co2         *measurement
	humidity    *measurement
	temperature *measurement

	mqtt mqttClient.Client

	homeassistantStatus = make(chan string, 10)
)

func main() {

	mqtt = connectMqtt()

	dev, err := openSCD30()
	if err != nil {
		log.Fatalf("error %v", err)
	}

	if config.ForceCalibration > 1 {
		forcedCalibration(dev)
		return
	}

	if err := dev.StartMeasurements(config.Interval); err != nil {
		log.Fatalf("error %v", err)
	}

	if err := dev.SetAutomaticSelfCalibration(config.AutoCalibration); err != nil {
		log.Fatalf("error %v", err)
	}

	announceMeToMqttDiscovery(mqtt)

	for {
		select {
		case <-time.After(time.Duration(config.Interval) * time.Second):
			checkMeasurement(dev)
		case status := <-homeassistantStatus:
			if status == "online" {
				// HA became online, send discovery so it knows about entities
				go announceMeToMqttDiscovery(mqtt)
			} else if status != "offline" {
				logInfo.Printf("unknown HA status message %s", status)
			}

		}
	}
}

func forcedCalibration(dev *scd30.SCD30) {
	err := dev.SetAutomaticSelfCalibration(0)
	if err := dev.StartMeasurements(2); err != nil {
		log.Fatalf("error %v", err)
	}
	fmt.Printf("Forced calibration %v ppm, waiting for 2 minutes\n", config.ForceCalibration)
	time.Sleep(125 * time.Second)
	err = dev.SetForcedCalibration(config.ForceCalibration)
	if err != nil {
		log.Fatalf("error %v", err)
	}
	fmt.Printf("Forced calibration %v ppm done\n", config.ForceCalibration)
}

func openSCD30() (*scd30.SCD30, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	bus, err := i2creg.Open("")
	if err != nil {
		return nil, err
	}

	dev, err := scd30.Open(bus)
	if err != nil {
		return nil, err
	}

	var to uint16 = config.TempOffset
	if o, err := dev.GetTemperatureOffset(); err != nil {
		return nil, err
	} else {
		logInfo.Printf("Got temp offset %d", o)
		if o != to {
			logInfo.Printf("Setting offset to %d", to)
			if err := dev.SetTemperatureOffset(to); err != nil {
				return nil, err
			}
		}
	}
	return dev, nil
}

func checkMeasurement(dev *scd30.SCD30) {
	if has, err := dev.HasMeasurement(); err != nil {
		log.Fatalf("error %v", err)
	} else if !has {
		return
	}

	m, err := dev.GetMeasurement()
	if err != nil {
		log.Fatalf("error %v", err)
	}

	logDebug.Printf("Got measure %f ppm %f%% %fC", m.CO2, m.Humidity, m.Temperature)

	publishIfNeeded(m.CO2, co2, 50)
	publishIfNeeded(m.Temperature, temperature, 0.5)
	publishIfNeeded(m.Humidity, humidity, 3)
}

func publishIfNeeded(current float32, old *measurement, diff float64) {
	if time.Since(old.time) < 600*time.Second && math.Abs(float64(current)-float64(old.value)) < diff {
		return
	}

	if current < old.min || current > old.max {
		logInfo.Printf("value for %s is out of range %f", old.id, current)
		return
	}

	old.time = time.Now()
	old.value = current

	publish(mqtt, stateTopic(old.id), fmt.Sprintf(old.format, current))
}

func subscribe(mqtt mqttClient.Client) {
	logInfo.Print("subscribed to topics")
	mqtt.Subscribe("homeassistant/status", 0, haStatusHandler)
}

func haStatusHandler(mqtt mqttClient.Client, msg mqttClient.Message) {
	body := string(msg.Payload())
	logInfo.Printf("received HA status %s", body)
	homeassistantStatus <- body
}

func announceMeToMqttDiscovery(mqtt mqttClient.Client) {
	publishDiscovery(mqtt, "co2", "co2", "ppm", "carbon_dioxide")
	publishDiscovery(mqtt, "temperature", "temperature", "°C", "temperature")
	publishDiscovery(mqtt, "humidity", "humidity", "%", "humidity")
}

func publishDiscovery(mqtt mqttClient.Client, id string, name string, unit string, class string) {
	uid := config.Id + "_" + id
	pname := config.Name + " " + name
	discoveryTopic := fmt.Sprintf("homeassistant/sensor/%s/config", uid)
	msg := discoveryMsg(id, uid, pname, unit, class)
	publish(mqtt, discoveryTopic, msg)
}

func publish(mqtt mqttClient.Client, topic string, msg interface{}) {

	logDebug.Printf("publish to %s: %s", topic, msg)

	t := mqtt.Publish(topic, 0, false, msg)
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			logInfo.Printf("publishing msg failed %v", t.Error())
		}
	}()
}

func discoveryMsg(id string, uid string, name string, unit string, class string) []byte {
	msg := make(map[string]interface{})
	msg["unique_id"] = uid
	msg["name"] = name

	dev := make(map[string]string)
	msg["device"] = dev
	dev["identifiers"] = config.Id
	dev["manufacturer"] = "Sensirion"
	dev["name"] = "Sensirion SCD30"
	dev["model"] = "SCD30"

	msg["state_topic"] = stateTopic(id)

	msg["expire_after"] = 1800

	msg["unit_of_measurement"] = unit
	msg["state_class"] = "measurement"
	msg["device_class"] = class

	jsonm, err := json.Marshal(msg)
	if err != nil {
		logInfo.Printf("cannot marshal json %v", err)
	}
	return jsonm
}

func stateTopic(id string) string {
	return "scd30/" + config.Id + "/" + id
}

func init() {

	co2 = &measurement{id: "co2", format: "%.0f", min: 100, max: 10000}
	humidity = &measurement{id: "humidity", format: "%.0f", min: 1, max: 100}
	temperature = &measurement{id: "temperature", format: "%.1f", min: -50, max: 150}

	err := envconfig.Process("scd30", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	logInfo = log.New(os.Stdout, "INFO  ", log.Ldate|log.Ltime|log.Lmsgprefix)

	if config.Debug {
		logDebug = log.New(os.Stdout, "DEBUG ", log.Ldate|log.Ltime|log.Lmsgprefix)
	} else {
		logDebug = log.New(ioutil.Discard, "DEBUG ", 0)
	}
}

func connectHandler(client mqttClient.Client) {
	options := client.OptionsReader()
	logInfo.Printf("MQTT connected to %s", options.Servers())
	subscribe(client)
}

func connectMqtt() mqttClient.Client {

	opts := mqttClient.NewClientOptions().
		AddBroker(config.MqttUrl).
		SetClientID(config.MqttClientId).
		SetOrderMatters(false).
		SetKeepAlive(150 * time.Second).
		SetAutoReconnect(true).
		SetOnConnectHandler(connectHandler)

	if len(config.MqttUser) > 0 {
		opts = opts.SetUsername(config.MqttUser)
	}

	if len(config.MqttPwd) > 0 {
		opts = opts.SetPassword(config.MqttPwd)
	}

	logInfo.Printf("connecting to mqtt %s client id %s user %s", opts.Servers, opts.ClientID, opts.Username)

	c := mqttClient.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return c
}
