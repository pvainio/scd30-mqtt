module github.com/pvainio/scd30-mqtt

go 1.17

require github.com/pvainio/scd30 v0.0.1

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/net v0.0.0-20211029224645-99673261e6eb // indirect
)

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/sigurn/crc8 v0.0.0-20160107002456-e55481d6f45c // indirect
	github.com/sigurn/utils v0.0.0-20190728110027-e1fefb11a144 // indirect
	periph.io/x/conn/v3 v3.6.9
	periph.io/x/host/v3 v3.7.1
)
