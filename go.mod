module github.com/pvainio/scd30-mqtt

go 1.20

require github.com/pvainio/scd30 v0.0.3

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/kelseyhightower/envconfig v1.4.0
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/host/v3 v3.7.2
)

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	github.com/sigurn/crc8 v0.0.0-20220107193325-2243fe600f9f // indirect
)