# Fannn

CLI tool to control Xiaomi Mi Smart Fan (dmaker.p18) via local miIO protocol.

## Install

```bash
go install github.com/u3mur4/xiaomi-fan-2@latest
```

## Config

Create `~/.config/xiaomi-fan-2/config.yaml`:

```yaml
location: "192.168.1.100"
id: 123456789
token: "yourdeviceToken"
```

Or use flags: `--location`, `--id`, `--token`

Or save with name: `xiaomi-fan-2 save myfan --name myfan --location 192.168.1.100 --id 123456789 --token yourToken`

## Usage

```bash
xiaomi-fan-2 on              # turn on
xiaomi-fan-2 off             # turn off
xiaomi-fan-2 toggle          # toggle power
xiaomi-fan-2 up              # increase speed
xiaomi-fan-2 down            # decrease speed
xiaomi-fan-2 next            # next speed level
xiaomi-fan-2 swing           # toggle horizontal swing
xiaomi-fan-2 angle [DEG]     # set/next swing angle
xiaomi-fan-2 mode            # toggle direct/natural breeze
xiaomi-fan-2 delay_off [MIN] # turn off after minutes
xiaomi-fan-2 status          # show status
```

## Bar Integration

```bash
xiaomi-fan-2 waybar          # waybar output
xiaomi-fan-2 polybar         # polybar output
```

## Server

```bash
xiaomi-fan-2 server          # HTTP server on port 35352
```

## Protocol

https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:fan:0000A005:dmaker-1c:1
