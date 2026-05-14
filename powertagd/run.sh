#!/bin/bash
set -e

OPTIONS="/data/options.json"

DEVICE=$(jq -r '.device // "/dev/ttyACM0"' "$OPTIONS")
BROKER=$(jq -r '.mqtt_broker // ""' "$OPTIONS")
USER=$(jq -r '.mqtt_user // ""' "$OPTIONS")
PASSWORD=$(jq -r '.mqtt_password // ""' "$OPTIONS")

# If broker is empty, try HA's Mosquitto service API
if [ -z "$BROKER" ]; then
    if MQTT_INFO=$(curl -s -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" http://supervisor/services/mqtt 2>/dev/null); then
        MQTT_HOST=$(echo "$MQTT_INFO" | jq -r '.data.host // empty')
        MQTT_PORT=$(echo "$MQTT_INFO" | jq -r '.data.port // empty')
        MQTT_USER=$(echo "$MQTT_INFO" | jq -r '.data.username // empty')
        MQTT_PASS=$(echo "$MQTT_INFO" | jq -r '.data.password // empty')
        if [ -n "$MQTT_HOST" ]; then
            BROKER="tcp://${MQTT_HOST}:${MQTT_PORT}"
            USER="$MQTT_USER"
            PASSWORD="$MQTT_PASS"
            echo "[INFO] Auto-detected Mosquitto broker at ${BROKER}"
        fi
    fi
fi

if [ -z "$BROKER" ]; then
    echo "[FATAL] No MQTT broker configured and Mosquitto add-on not found"
    exit 1
fi

echo "[INFO] Starting powertagd on ${DEVICE}"
echo "[INFO] MQTT broker: ${BROKER}"

# Build mqtt args
MQTT_ARGS="-broker ${BROKER}"
if [ -n "$USER" ]; then
    MQTT_ARGS="${MQTT_ARGS} -user ${USER} -password ${PASSWORD}"
fi

# Run powertagd piped to powertag2mqtt
exec /usr/bin/powertagd -d "${DEVICE}" 2>&1 | /usr/bin/powertag2mqtt ${MQTT_ARGS}
