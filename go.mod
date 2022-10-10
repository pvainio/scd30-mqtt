module github.com/pvainio/scd30-mqtt

go 1.19

require github.com/pvainio/scd30 v0.0.3

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	golang.org/x/net v0.0.0-20221004154528-8021a29435af // indirect
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0 // indirect
)

require (
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/sigurn/crc8 v0.0.0-20220107193325-2243fe600f9f // indirect
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/host/v3 v3.7.2
)
