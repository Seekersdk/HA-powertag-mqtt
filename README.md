# PowerTagd Home Assistant Add-on Repository

Home Assistant add-on for Schneider Electric PowerTag energy monitors.

## Add-on: PowerTagd

Reads data from PowerTag sensors via a Zigbee USB dongle (Sonoff Zigbee 3.0 USB Dongle Plus V2) and publishes to MQTT with full Home Assistant auto-discovery.

### Installation

1. Add this repository to Home Assistant:
   **Settings > Add-ons > Add-on Store > Menu (three dots) > Repositories**
   and paste the URL of this repository.

2. Install the **PowerTagd** add-on.

3. Connect your Zigbee USB dongle to the HAOS machine.

4. Configure the add-on:
   - **device**: USB device path (default `/dev/ttyACM0`, check Settings > System > Hardware)
   - **mqtt_broker**: Leave empty to auto-detect Mosquitto add-on, or set manually (e.g. `tcp://192.168.1.10:1883`)
   - **mqtt_user** / **mqtt_password**: Leave empty for auto-detection from Mosquitto add-on

5. Start the add-on. PowerTag sensors will appear automatically in Home Assistant via MQTT discovery.

### Components

- **powertagd** - C daemon that communicates with PowerTag sensors over Zigbee (based on [fdamm/powertagd](https://github.com/fdamm/powertagd) fork)
- **powertag2mqtt** - Go bridge that parses powertagd output and publishes to MQTT with HA discovery
