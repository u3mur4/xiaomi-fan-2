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
	Model         string
	FirmwareVer   string
	HardwareVer   string
	MCUFirmwareVer string
	MAC           string
	Life          int
}

type FanStatus struct {
	Power      bool
	Level      FanLevel
	Mode       Mode
	Swing      bool
	Angle      HorizontalAngle
	Brightness bool
	Alarm      bool
	SpeedLevel uint8
	ChildLock  bool
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

func (fan *Fan) GetDeviceInformation() (*DeviceInformation, error) {
	cmd := &command{
		ID:     fan.nextMsgID(),
		Method: method("miIO.info"),
		Params: []any{},
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return nil, err
	}

	result, _, _, err := jsonparser.Get(response, "result")
	if err != nil {
		return nil, fmt.Errorf("miIO.info missing result: %w", err)
	}

	var info struct {
		Model         string `json:"model"`
		FirmwareVer   string `json:"fw_ver"`
		HardwareVer   string `json:"hw_ver"`
		MCUFirmwareVer string `json:"mcu_fw_ver"`
		MAC           string `json:"mac"`
		Life          int    `json:"life"`
	}
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, fmt.Errorf("parse miIO.info result: %w", err)
	}

	return &DeviceInformation{
		Model:          info.Model,
		FirmwareVer:    info.FirmwareVer,
		HardwareVer:    info.HardwareVer,
		MCUFirmwareVer: info.MCUFirmwareVer,
		MAC:            info.MAC,
		Life:           info.Life,
	}, nil
}

func (fan *Fan) GetAllStatus() (*FanStatus, error) {
	cmd, err := fan.createCommand(getProperties,
		switchStatus(nil),
		fanLevel(nil),
		mode(nil),
		horizontalSwing(nil),
		horizontalAngle(nil),
		brightness(nil),
		alarm(nil),
		speedLevel(nil),
		childLock(nil),
	)
	if err != nil {
		return nil, err
	}

	response, err := fan.SendPayloadAndWait(cmd)
	if err != nil {
		return nil, err
	}

	vals := make([]any, 9)
	for i := range vals {
		vals[i], err = fan.getResponseValue(response, i)
		if err != nil {
			return nil, fmt.Errorf("property %d: %w", i, err)
		}
	}

	return &FanStatus{
		Power:      vals[0].(bool),
		Level:      FanLevel(vals[1].(int64)),
		Mode:       Mode(vals[2].(int64)),
		Swing:      vals[3].(bool),
		Angle:      HorizontalAngle(vals[4].(int64)),
		Brightness: vals[5].(bool),
		Alarm:      vals[6].(bool),
		SpeedLevel: uint8(vals[7].(int64)),
		ChildLock:  vals[8].(bool),
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

func (fan *Fan) SetBrightness(status bool) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, brightness(&status))
	if err != nil {
		return err
	}
	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetBrightness() (bool, error) {
	cmd, err := fan.createCommand(getProperties, brightness(nil))
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

func (fan *Fan) SetAlarm(status bool) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, alarm(&status))
	if err != nil {
		return err
	}
	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetAlarm() (bool, error) {
	cmd, err := fan.createCommand(getProperties, alarm(nil))
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

func (fan *Fan) SetMotorControl(val uint8) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, motorControl(&val))
	if err != nil {
		return err
	}
	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) SetSpeedLevel(level uint8) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, speedLevel(&level))
	if err != nil {
		return err
	}
	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetSpeedLevel() (uint8, error) {
	cmd, err := fan.createCommand(getProperties, speedLevel(nil))
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
	return uint8(val.(int64)), nil
}

func (fan *Fan) SetChildLock(status bool) error {
	defer fan.changed()
	cmd, err := fan.createCommand(setProperties, childLock(&status))
	if err != nil {
		return err
	}
	_, err = fan.SendPayloadAndWait(cmd)
	return err
}

func (fan *Fan) GetChildLock() (bool, error) {
	cmd, err := fan.createCommand(getProperties, childLock(nil))
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

func (fan *Fan) drain() {
	fan.connection.SetReadDeadline(time.Now().Add(time.Millisecond))
	buf := make([]byte, 1024*4)
	fan.connection.Read(buf)
}

func (fan *Fan) sendAndReceive(v any) (b []byte, err error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	msgID, err := jsonparser.GetInt(payload, "id")
	if err != nil {
		return nil, err
	}

	if fan.timeout > 0 {
		fan.connection.SetWriteDeadline(time.Now().Add(fan.timeout))
	}

	_, err = fan.connection.Write(payload)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024*4)
	fan.connection.SetReadDeadline(time.Now().Add(fan.timeout))
	n, err := fan.connection.Read(buf)
	if err != nil {
		fan.drain()
		return nil, fmt.Errorf("timeout for response")
	}

	respID, err := jsonparser.GetInt(buf[:n], "id")
	if err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if respID != msgID {
		return nil, fmt.Errorf("unexpected response id %d, expected %d", respID, msgID)
	}

	code, err := jsonparser.GetInt(buf[:n], "error", "code")
	if err == nil && code != 0 {
		msg, _ := jsonparser.GetString(buf[:n], "error.message")
		return nil, fmt.Errorf("invalid packet error code %d:%s", code, msg)
	}

	return buf[:n], nil
}

func (fan *Fan) SendPayloadAndWait(v any) (response []byte, err error) {
	return fan.sendAndReceive(v)
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
	}

	fan.Timeout(500 * time.Millisecond)

	return fan, nil
}
