# Fannn

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

Set a default device (so you can omit `--device`):

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
fan2ctl --device bedroom status     # use a different device
fan2ctl --id X --location Y --token Z on  # no config needed
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
```

## Bar Integration

```bash
fan2ctl waybar              # waybar output (with online/offline indicator)
fan2ctl polybar             # polybar output (with online/offline indicator)
```

Notifications from CLI commands (e.g. `fan2ctl on`) automatically refresh the widget output.

## Server

```bash
fan2ctl server              # HTTP server on port 35352
```

## Docker

```bash
docker build -t fan2ctl .
docker run -p 35352:35352 fan2ctl
```

## Protocol

https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:fan:0000A005:dmaker-1c:1
