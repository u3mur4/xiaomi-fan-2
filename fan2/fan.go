package fan2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/buger/jsonparser"
	"github.com/u3mur4/xiaomi-fan-2/miio"
)

type DeviceInformation struct {
	Manufacturer   string
	Model          string
	SerialNumber   string
	FirmwareVerson string
}

type Mode uint8

const (
	StraightWind Mode = 0
	Sleep        Mode = 1
)

func (m Mode) Toggle() Mode {
	if m == StraightWind {
		return Sleep
	}
	return StraightWind
}

type Fan struct {
	responses  map[int64]chan []byte
	connection net.Conn
	deviceID   string
	debug      io.Writer
	timeout    time.Duration
	notify     func()
}

func (fan *Fan) OnChange(fn func()) {
	fan.notify = fn
}

func (fan *Fan) changed() {
	if fan.notify != nil {
		fan.notify()
	}
}

func (fan *Fan) Debug(output io.Writer) {
	miio.Debug(fan.connection, output)
	fan.debug = output
}

func (fan *Fan) Timeout(duration time.Duration) {
	fan.timeout = duration
}

func (fan *Fan) debugMsg(msg string) {
	fmt.Fprintln(fan.debug, msg)
}

func (fan *Fan) reader() {
	response := make([]byte, 1024*4)
	for {
		if fan.timeout > 0 {
			fan.connection.SetReadDeadline(time.Now().Add(fan.timeout))
		}
		n, err := fan.connection.Read(response)
		if err != nil {
			if err, ok := err.(net.Error); ok && !err.Timeout() {
				if fan.debug != nil {
					fan.debugMsg(fmt.Sprintf("<-error: %s", err))
				}
			return
			}
		}

		msgID, err := jsonparser.GetInt(response[:n], "id")
		if err != nil {
			continue
		}

		if ch, ok := fan.responses[msgID]; ok {
			ch <- response[:n]
		}

		delete(fan.responses, msgID)

	}
}

func (fan *Fan) GetDeviceInformation() (*DeviceInformation, error) {
	cmd, err := fan.createCommand(getProperties, deviceManufacturer(), deviceModel(), deviceSerialNumber(), deviceFirmwareVersion())
	if err != nil {
		return nil, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return nil, err
	}

	manufacturer, err := fan.getResponseValue(response, 0)
	if err != nil {
		return nil, err
	}

	model, err := fan.getResponseValue(response, 1)
	if err != nil {
		return nil, err
	}

	serialNumber, err := fan.getResponseValue(response, 2)
	if err != nil {
		return nil, err
	}

	firmwareVersion, err := fan.getResponseValue(response, 3)
	if err != nil {
		return nil, err
	}

	return &DeviceInformation{
		Manufacturer:   manufacturer.(string),
		Model:          model.(string),
		SerialNumber:   serialNumber.(string),
		FirmwareVerson: firmwareVersion.(string),
	}, nil
}

func (fan *Fan) SetLevel(level FanLevel) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, fanLevel(&level))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetLevel() (FanLevel, error) {
	cmd, err := fan.createCommand(getProperties, fanLevel(nil))
	if err != nil {
		return 0, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return 0, err
	}

	val, err := fan.getResponseValue(response, 0)
	if err != nil {
		return 0, err
	}

	return FanLevel(val.(int64)), nil
}

func (fan *Fan) GetMode() (Mode, error) {
	cmd, err := fan.createCommand(getProperties, mode(nil))
	if err != nil {
		return 0, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return 0, err
	}

	val, err := fan.getResponseValue(response, 0)
	if err != nil {
		return 0, err
	}

	return Mode(val.(int64)), nil
}

func (fan *Fan) SetMode(mode_ Mode) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, mode(&mode_))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) getResponseValue(response []byte, index int) (any, error) {
	indexKey := fmt.Sprintf("[%d]", index)

	code, err := jsonparser.GetInt(response, "result", indexKey, "code")
	if err == nil && code != 0 {
		errMsg := "Unknown error"
		switch code {
		case 1:
			errMsg = "The request was received, but the operation has not completed"
		case -4001:
			errMsg = "Unreadable attributes"
		case -4002:
			errMsg = "Property is not writable"
		case -4003:
			errMsg = "Properties, methods, events do not exist"
		case -4004:
			errMsg = "Other internal errors"
		case -4005:
			errMsg = "Attribute value error"
		case -4006:
			errMsg = "Method in parameter error"
		case -4007:
			errMsg = "did error"
		}
		return nil, fmt.Errorf("invalid message error code %d: %s", code, errMsg)
	}

	value, dataType, _, err := jsonparser.Get(response, "result", indexKey, "value")
	if err != nil {
		return nil, err
	}

	switch dataType {
	case jsonparser.String:
		return jsonparser.ParseString(value)
	case jsonparser.Number:
		return jsonparser.ParseInt(value)
	case jsonparser.Boolean:
		return jsonparser.ParseBoolean(value)
	default:
		return nil, fmt.Errorf("not supported value type")
	}

}

func (fan *Fan) SetHorizontalAngle(status HorizontalAngle) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, horizontalAngle(&status))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetHorizontalAngle() (HorizontalAngle, error) {
	cmd, err := fan.createCommand(getProperties, horizontalAngle(nil))
	if err != nil {
		return HorizontalAngle30, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return HorizontalAngle30, err
	}

	val, err := fan.getResponseValue(response, 0)
	if err != nil {
		return HorizontalAngle30, err
	}

	return HorizontalAngle(val.(int64)), nil
}

func (fan *Fan) SetHorizontalSwing(status bool) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, horizontalSwing(&status))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetHorizontalSwing() (bool, error) {
	cmd, err := fan.createCommand(getProperties, horizontalSwing(nil))
	if err != nil {
		return false, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return false, err
	}

	val, err := fan.getResponseValue(response, 0)
	if err != nil {
		return false, err
	}

	return val.(bool), nil
}

func (fan *Fan) GetPower() (bool, error) {
	cmd, err := fan.createCommand(getProperties, switchStatus(nil))
	if err != nil {
		return false, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return false, err
	}

	val, err := fan.getResponseValue(response, 0)
	if err != nil {
		return false, err
	}

	return val.(bool), nil
}

func (fan *Fan) DelayOff(minutes int64) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, delayOff(&minutes))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) On() error {
	defer fan.changed()
	status := true
	cmd, err := fan.createCommand(setProperties, switchStatus(&status))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) Off() error {
	defer fan.changed()
	status := false
	cmd, err := fan.createCommand(setProperties, switchStatus(&status))
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) Toggle() error {
	defer fan.changed()
	power, err := fan.GetPower()
	if err != nil {
		return err
	}
	if power {
		return fan.Off()
	}
	return fan.On()
}

func (fan *Fan) createCommand(method method, params ...*param) (cmd *command, err error) {
	cmd = &command{
		ID:     fan.nextMsgID(),
		Method: method,
		Params: nil,
	}

	// fill missing deviceID
	for _, param := range params {
		param.DeviceID = fan.deviceID
	}

	switch method {
	// for events and actions the params is an object not an array
	case eventOccured:
		fallthrough
	case action:
		if len(params) != 1 {
			return nil, fmt.Errorf("%s method can contain only one param object, but got %d", method, len(params))
		}
		cmd.Params = params[0]
		// fallthrough
	default:
		paramArray := make([]any, 0, len(params))
		for _, param := range params {
			paramArray = append(paramArray, param)
		}
		cmd.Params = paramArray
	}

	return cmd, nil
}

func (fan *Fan) SendPayloadJSON(payload []byte, waitForRespone bool) (response []byte, err error) {
	msgID, err := jsonparser.GetInt(payload, "id")
	if err != nil {
		return nil, err
	}

	if waitForRespone {
		// register channel before sending payload
		fan.responses[msgID] = make(chan []byte)
		defer func() {
			delete(fan.responses, msgID)
		}()
	}

	if fan.timeout > 0 {
		fan.connection.SetWriteDeadline(time.Now().Add(fan.timeout))
	}

	_, err = fan.connection.Write(payload)
	if err != nil {
		return nil, err
	}

	if !waitForRespone {
		return nil, nil
	}

	select {
	case response = <-fan.responses[msgID]:
		if len(response) == 0 || response == nil {
			return nil, fmt.Errorf("no response")
		}
		return response, nil
	case <-time.Tick(500 * time.Millisecond):
		return nil, fmt.Errorf("timeout for response")
	}

}

func (fan *Fan) SendPayloadAndWait(v any) (response []byte, err error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	response, err = fan.SendPayloadJSON(payload, true)
	if err != nil {
		return nil, err
	}

	code, err := jsonparser.GetInt(response, "error", "code")
	if err == nil && code != 0 {
		msg, _ := jsonparser.GetString(response, "error.message")
		return nil, fmt.Errorf("invalid packet error code %d:%s", code, msg)
	}

	return response, nil
}

func (fan *Fan) SendPayload(v any) (err error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = fan.SendPayloadJSON(payload, false)
	if err != nil {
		return err
	}

	return nil
}

func (fan *Fan) Close() error {
	return fan.connection.Close()
}

var id int32 = 0

func (fan *Fan) nextMsgID() int32 {
	if id == 0 {
		id = rand.Int31()
	}
	id++
	return id
}

func NewFan2(ip string, deviceID uint32, deciveToken string) (*Fan, error) {
	rand.Seed(time.Now().UnixNano())
	token, err := hex.DecodeString(deciveToken)
	if err != nil {
		return nil, err
	}

	connection, err := miio.Dial(ip, deviceID, token)
	if err != nil {
		return nil, err
	}

	fan := &Fan{
		connection: connection,
		deviceID:   fmt.Sprintf("%d", deviceID),
		responses:  make(map[int64]chan []byte),
	}

	go fan.reader()
	fan.Timeout(time.Second * 2)

	return fan, nil
}
