#!/bin/bash
set -e

OPTIONS="/data/options.json"

echo "[DEBUG] Options file contents:"
cat "$OPTIONS"

DEVICE=$(jq -r '.device // "/dev/ttyACM0"' "$OPTIONS")
BROKER=$(jq -r '.mqtt_broker // ""' "$OPTIONS")
USER=$(jq -r '.mqtt_user // ""' "$OPTIONS")
PASSWORD=$(jq -r '.mqtt_password // ""' "$OPTIONS")

echo "[DEBUG] DEVICE=${DEVICE} BROKER=${BROKER}"
echo "[DEBUG] SUPERVISOR_TOKEN set: $([ -n "$SUPERVISOR_TOKEN" ] && echo 'yes' || echo 'no')"

# If broker is empty, try HA's Mosquitto service API
if [ -z "$BROKER" ]; then
    echo "[DEBUG] Trying Supervisor API..."
    MQTT_INFO=$(curl -sv -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" http://supervisor/services/mqtt 2>&1) || true
    echo "[DEBUG] Supervisor response: ${MQTT_INFO}"

    MQTT_HOST=$(echo "$MQTT_INFO" | grep -v '^\*\|^>\|^<\|^{' | jq -r '.data.host // empty' 2>/dev/null || true)
    if [ -z "$MQTT_HOST" ]; then
        MQTT_HOST=$(echo "$MQTT_INFO" | jq -r '.data.host // empty' 2>/dev/null || true)
    fi
    MQTT_PORT=$(echo "$MQTT_INFO" | jq -r '.data.port // empty' 2>/dev/null || true)
    MQTT_USER=$(echo "$MQTT_INFO" | jq -r '.data.username // empty' 2>/dev/null || true)
    MQTT_PASS=$(echo "$MQTT_INFO" | jq -r '.data.password // empty' 2>/dev/null || true)

    echo "[DEBUG] MQTT_HOST=${MQTT_HOST} MQTT_PORT=${MQTT_PORT}"

    if [ -n "$MQTT_HOST" ]; then
        BROKER="tcp://${MQTT_HOST}:${MQTT_PORT}"
        USER="$MQTT_USER"
        PASSWORD="$MQTT_PASS"
        echo "[INFO] Auto-detected Mosquitto broker at ${BROKER}"
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
