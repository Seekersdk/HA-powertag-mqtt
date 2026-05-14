#!/usr/bin/with-contenv bashio

# Read config from HA add-on options
DEVICE=$(bashio::config 'device')
BROKER=$(bashio::config 'mqtt_broker')
USER=$(bashio::config 'mqtt_user')
PASSWORD=$(bashio::config 'mqtt_password')

# If broker is empty, try to use HA's built-in Mosquitto
if bashio::var.is_empty "${BROKER}"; then
    if bashio::services.available "mqtt"; then
        BROKER="tcp://$(bashio::services mqtt 'host'):$(bashio::services mqtt 'port')"
        USER="$(bashio::services mqtt 'username')"
        PASSWORD="$(bashio::services mqtt 'password')"
        bashio::log.info "Using HA Mosquitto broker at ${BROKER}"
    else
        bashio::log.fatal "No MQTT broker configured and Mosquitto add-on not found"
        exit 1
    fi
fi

bashio::log.info "Starting powertagd on ${DEVICE}"
bashio::log.info "MQTT broker: ${BROKER}"

# Build mqtt args
MQTT_ARGS="-broker ${BROKER}"
if bashio::var.has_value "${USER}"; then
    MQTT_ARGS="${MQTT_ARGS} -user ${USER} -password ${PASSWORD}"
fi

# Run powertagd piped to powertag2mqtt
exec /usr/bin/powertagd -d "${DEVICE}" | /usr/bin/powertag2mqtt ${MQTT_ARGS}
