# xiaomi-fan-2

CLI tool to control Xiaomi Mi Smart Fan (dmaker.p18) via local miIO protocol.

## Install

```bash
go install github.com/u3mur4/xiaomi-fan-2/cli/fan2ctl@latest
```

Or build from source:

```bash
git clone https://github.com/u3mur4/xiaomi-fan-2
cd xiaomi-fan-2
go build -o fan2ctl ./cli/fan2ctl/
```

## Device Config

Save fan connection details as named devices:

```bash
fan2ctl device add living_room --location 192.168.1.100 --id 123456789 --token yourToken
fan2ctl device add bedroom       --location 192.168.1.101 --id 987654321 --token anotherToken
```

If you save only one device, it's used automatically — no need for `--device` or `device default`.

Set a default device (or let the single-device auto-detect handle it):

```bash
fan2ctl device default living_room
```

List and remove devices:

```bash
fan2ctl device list
fan2ctl device remove bedroom
```

Devices are stored in `~/.config/xiaomi-fan-2/config.yaml`:

```yaml
_default: living_room
living_room:
  location: "192.168.1.100"
  id: 123456789
  token: "yourToken"
bedroom:
  location: "192.168.1.101"
  id: 987654321
  token: "anotherToken"
```

### Override device or use inline flags

```bash
fan2ctl --device bedroom on          # use a specific device (override default)
fan2ctl --id X --location Y --token Z on   # no config needed, inline connection
```

## Usage

```bash
fan2ctl online              # check if fan is reachable
fan2ctl on                  # turn on
fan2ctl off                 # turn off
fan2ctl toggle              # toggle power
fan2ctl up                  # increase speed
fan2ctl down                # decrease speed
fan2ctl next                # next speed level
fan2ctl swing               # toggle horizontal swing
fan2ctl angle [DEG]         # set/next swing angle
fan2ctl mode                # toggle direct/natural breeze
fan2ctl delay_off [MIN]     # turn off after minutes
fan2ctl status              # show fan state + online status
fan2ctl api                 # start REST API server
fan2ctl web                 # start web UI
```

## Bar Integration

```bash
fan2ctl waybar              # waybar output (with online/offline indicator)
fan2ctl polybar             # polybar output (with online/offline indicator)
```

Notifications from CLI commands (e.g. `fan2ctl on`) automatically refresh the widget output.

### Waybar config example

```json
"custom/xiaomi-fan": {
  "restart-interval": 5,
  "exec": "fan2ctl waybar",
  "on-click": "fan2ctl swing",
  "on-click-right": "fan2ctl toggle",
  "on-click-middle": "fan2ctl mode",
  "on-scroll-up": "fan2ctl up",
  "on-scroll-down": "fan2ctl down"
}
```

## API

```bash
fan2ctl api                 # REST API on port 35352
fan2ctl api --port 9090     # custom port
fan2ctl api --origin http://localhost:8080  # allow CORS from web UI
```

Endpoints:

```
POST /api/command  {"cmd":"toggle", "device":"living_room"}
GET  /api/status?device=living_room
```

## Web UI

```bash
fan2ctl web                 # web UI on port 8080 (includes built-in API)
fan2ctl web --port 9090     # custom port
```

## Docker

```bash
docker build -t fan2ctl .
docker run -p 8080:8080 fan2ctl web
```

## Protocol

https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:fan:0000A005:dmaker-1c:1
