package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const ProgNameMqtt string = "powertag2mqtt"

type SensorMeta struct {
	Unit        string
	DeviceClass string
	StateClass  string
}

var sensorMetaMap = map[string]SensorMeta{
	// Power (W)
	"power_p1_active":     {Unit: "W", DeviceClass: "power", StateClass: "measurement"},
	"power_p2_active":     {Unit: "W", DeviceClass: "power", StateClass: "measurement"},
	"power_p3_active":     {Unit: "W", DeviceClass: "power", StateClass: "measurement"},
	"power_p1_reactive":   {Unit: "VAR", DeviceClass: "reactive_power", StateClass: "measurement"},
	"power_p2_reactive":   {Unit: "VAR", DeviceClass: "reactive_power", StateClass: "measurement"},
	"power_p3_reactive":   {Unit: "VAR", DeviceClass: "reactive_power", StateClass: "measurement"},
	"power_p1_apparent":   {Unit: "VA", DeviceClass: "apparent_power", StateClass: "measurement"},
	"power_p2_apparent":   {Unit: "VA", DeviceClass: "apparent_power", StateClass: "measurement"},
	"power_p3_apparent":   {Unit: "VA", DeviceClass: "apparent_power", StateClass: "measurement"},
	"total_power_active":  {Unit: "W", DeviceClass: "power", StateClass: "measurement"},
	"total_power_apparent": {Unit: "VA", DeviceClass: "apparent_power", StateClass: "measurement"},
	"total_power_reactive": {Unit: "VAR", DeviceClass: "reactive_power", StateClass: "measurement"},

	// Voltage (V)
	"voltage_p1":       {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},
	"voltage_p2":       {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},
	"voltage_p3":       {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},
	"voltage_phase_ab": {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},
	"voltage_phase_bc": {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},
	"voltage_phase_ac": {Unit: "V", DeviceClass: "voltage", StateClass: "measurement"},

	// Current (A)
	"current_p1":      {Unit: "A", DeviceClass: "current", StateClass: "measurement"},
	"current_p2":      {Unit: "A", DeviceClass: "current", StateClass: "measurement"},
	"current_p3":      {Unit: "A", DeviceClass: "current", StateClass: "measurement"},
	"current_neutral": {Unit: "A", DeviceClass: "current", StateClass: "measurement"},

	// Energy delivered (kWh) - old naming (original powertagd)
	"energy_delivered": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p1_tx":     {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p2_tx":     {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p3_tx":     {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},

	// Energy delivered (kWh) - new naming (fdamm fork)
	"total_energy_tx":      {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p1_tx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p2_tx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p3_tx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_tx":    {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p1_tx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p2_tx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p3_tx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},

	// Energy received (kWh) - old naming
	"energy_received": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p1_rx":    {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p2_rx":    {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"energy_p3_rx":    {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},

	// Energy received (kWh) - new naming
	"total_energy_rx":      {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p1_rx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p2_rx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"total_energy_p3_rx":   {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_rx":    {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p1_rx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p2_rx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},
	"partial_energy_p3_rx": {Unit: "kWh", DeviceClass: "energy", StateClass: "total_increasing"},

	// Frequency (Hz)
	"freq": {Unit: "Hz", DeviceClass: "frequency", StateClass: "measurement"},

	// Power factor (%)
	"power_factor":    {Unit: "%", DeviceClass: "power_factor", StateClass: "measurement"},
	"power_factor_p1": {Unit: "%", DeviceClass: "power_factor", StateClass: "measurement"},
	"power_factor_p2": {Unit: "%", DeviceClass: "power_factor", StateClass: "measurement"},
	"power_factor_p3": {Unit: "%", DeviceClass: "power_factor", StateClass: "measurement"},

	// Breaker capacity (A, no device_class)
	"breaker_capacity": {Unit: "A"},
}

// Sensors to skip publishing discovery for (internal/multiplier fields)
var skipDiscovery = map[string]bool{
	"unit_of_measure":   true,
	"multiplier":        true,
	"divisor":           true,
	"freq_multiplier":   true,
	"freq_divisor":      true,
	"voltage_multiplier": true,
	"voltage_divisor":   true,
	"current_multiplier": true,
	"current_divisor":   true,
	"ac_power_multiplier": true,
	"ac_power_divisor":  true,
	"power_multiplier":  true,
	"power_divisor":     true,
	"measurement_type":  true,
	"freq_min":          true,
	"freq_max":          true,
	"voltage_p1_min":    true,
	"voltage_p1_max":    true,
	"voltage_p2_min":    true,
	"voltage_p2_max":    true,
	"voltage_p3_min":    true,
	"voltage_p3_max":    true,
	"current_p1_min":    true,
	"current_p1_max":    true,
	"current_p2_min":    true,
	"current_p2_max":    true,
	"current_p3_min":    true,
	"current_p3_max":    true,
	"power_p1_active_min": true,
	"power_p1_active_max": true,
	"power_p2_active_min": true,
	"power_p2_active_max": true,
	"power_p3_active_min": true,
	"power_p3_active_max": true,
	"line_current":      true,
	"line_current_p2":   true,
	"line_current_p3":   true,
	"current_p1_active": true,
	"current_p2_active": true,
	"current_p3_active": true,
	"current_p1_reactive": true,
	"current_p2_reactive": true,
	"current_p3_reactive": true,
	"zcl_ver":           true,
	"app_ver":           true,
	"stack_ver":         true,
	"date":              true,
	"sw_build":          true,
	"brand":             true,
	"mfr":               true,
	// Device info strings - appear in one-time reports, not regular measurement packets
	"model":             true,
	"fw_ver":            true,
	"hw_ver":            true,
	"serial":            true,
	"alarm":             true,
	"mount_position":    true,
}

var (
	publishedMu sync.Mutex
	published   = make(map[string]bool) // "deviceId:sensor" -> true
)

func sanitizeID(id string) string {
	// Replace characters not safe for MQTT topic/HA unique_id
	r := strings.NewReplacer(":", "_", "/", "_", " ", "_")
	return r.Replace(id)
}

func publishDiscovery(client mqtt.Client, deviceID string, sensor string) {
	key := deviceID + ":" + sensor
	publishedMu.Lock()
	if published[key] {
		publishedMu.Unlock()
		return
	}
	published[key] = true
	publishedMu.Unlock()

	if skipDiscovery[sensor] {
		return
	}

	// Skip raw hex register keys (e.g. "0x4502") - unknown ZCL attributes
	if strings.HasPrefix(sensor, "0x") {
		return
	}

	safeID := sanitizeID(deviceID)
	uniqueID := safeID + "_" + sensor
	stateTopic := "powertag/" + deviceID
	discoveryTopic := "homeassistant/sensor/powertag_" + uniqueID + "/config"

	meta, known := sensorMetaMap[sensor]

	// Use only the sensor field name - HA prepends device name automatically
	name := sensor

	config := map[string]interface{}{
		"name":           name,
		"object_id":      safeID + "_" + sensor,
		"state_topic":    stateTopic,
		"value_template": "{{ value_json." + sensor + " }}",
		"unique_id":      uniqueID,
		"device": map[string]interface{}{
			"identifiers":  []string{"powertag_" + safeID},
			"name":         "PowerTag " + deviceID,
			"manufacturer": "Schneider Electric",
		},
	}

	if known {
		if meta.Unit != "" {
			config["unit_of_measurement"] = meta.Unit
		}
		if meta.DeviceClass != "" {
			config["device_class"] = meta.DeviceClass
		}
		if meta.StateClass != "" {
			config["state_class"] = meta.StateClass
		}
	}

	payload, err := json.Marshal(config)
	if err != nil {
		log.Printf("discovery: failed to marshal config for %s/%s: %v", deviceID, sensor, err)
		return
	}

	token := client.Publish(discoveryTopic, 0, true, payload)
	token.Wait()
	log.Printf("discovery: published %s", discoveryTopic)
}

func main() {
	var broker, user, password string

	flag.StringVar(&broker, "broker", "tcp://192.168.0.20:1883", "MQTT broker URL")
	flag.StringVar(&user, "user", "", "MQTT username")
	flag.StringVar(&password, "password", "", "MQTT password")
	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprintf(os.Stderr, "%s: no data on stdin\n", ProgNameMqtt)
		fmt.Fprintf(os.Stderr, "%s expects data to be piped to stdin, i.e.:\n", ProgNameMqtt)
		fmt.Fprintf(os.Stderr, "    powertagd -d /dev/ttyACM0 | powertag2mqtt -broker tcp://host:1883\n")
		os.Exit(2)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(ProgNameMqtt).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second)

	if user != "" {
		opts.SetUsername(user)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect failed: %v", token.Error())
	}

	fmt.Printf("%s: connected to %s\n", ProgNameMqtt, broker)

	lnscan := bufio.NewScanner(os.Stdin)
	for lnscan.Scan() {
		line := lnscan.Text()
		if !strings.HasPrefix(line, "powertag,") {
			continue
		}

		sanitized := strings.Replace(line, "powertag,", "", 1)
		parts := strings.Split(sanitized, " ")
		if len(parts) < 2 {
			continue
		}

		tags := asMap(parts[0])
		measures := asMap(parts[1])

		deviceID, ok := tags["id"]
		if !ok {
			continue
		}

		// Publish discovery for any new sensors seen on this device
		for sensor := range measures {
			publishDiscovery(client, deviceID, sensor)
		}

		// Publish the measurement payload
		jsonStr, _ := json.Marshal(measures)
		token := client.Publish("powertag/"+deviceID, 0, false, jsonStr)
		token.Wait()
	}
}

func asMap(inputString string) map[string]string {
	m := make(map[string]string)
	for _, s := range strings.Split(inputString, ",") {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
