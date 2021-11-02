# Sensirion SCD30 CO2 sensor MQTT gateway for Home Assistant

## Overview

This gateway can be used to publish measurements SCD30 to mqtt.

It supports Home Assistant MQTT Discovery but can also be used without Home Assistant.

Only requirement is MQTT Broker to connect to.

## Example usecase

Attach SCD30 to Raspberry Pi Zero W I2C bus and run this gateway to publish CO2, temperature and humidity to Home Assistant.

## Configuration

Application is configure with environment variables

| variable        | required | default | description |
|-----------------|:--------:|---------|-------------|
| SCD30_MQTT_URL        |    x     |         | mqtt url, for example tcp://10.1.2.3:8883 |
| SCD30_MQTT_USER       |          |         | mqtt username |
| SCD30_MQTT_PASSWORD   |          |         | mqtt password |
| SCD30_MQTT_CLIENT_ID  |          | scd30   | mqtt client id |
| SCD30_DEBUG           |          | false   | enable debug output, true/false |
| SCD30_ID              |          | scd30   | home assistant discovery id |
| SCD30_TEMP_OFFSET     |          | 150     | temperature compensation offset |
| SCD30_NAME            |          | SCD30   | home assistant device name |
| SCD30_INTEVAL         |          | 50      | measurement interval in seconds |

## Usage

For example with following script
```sh
#!/bin/sh

# Change to your real mqtt url
export MQTT_URL=tcp://localhost:8883

./scd30-mqtt
```

## MQTT Topics used

- homeassistant/status subscribe to HA status changes
- scd30/_id_/co2 publish co2
- scd30/_id_/temperature publish temperature
- scd30/_id_/humidity publish humidity
