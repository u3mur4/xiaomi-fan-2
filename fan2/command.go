package fan2

type method string

const (
	setProperties     method = "set_properties"
	getProperties     method = "get_properties"
	action            method = "action"
	propertiesChanged method = "properties_changed"
	eventOccured      method = "event_occured"
)

type command struct {
	ID     int32       `json:"id"`
	Method method      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type param struct {
	DeviceID string `json:"did"`

	ServiceID uint32 `json:"siid"`

	PropertyID uint32 `json:"piid,omitempty"`
	ActionID   uint32 `json:"aiid,omitempty"`
	EventID    uint32 `json:"eiid,omitempty"`

	In interface{} `json:"in,omitempty"`

	Value     interface{} `json:"value,omitempty"`
	Arguments interface{} `json:"arguments,omitempty"`
}

func horizontalSwing(status *bool) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 4,
		Value:      status,
	}
}

func delayOff(minutes *int64) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 10,
		Value:      minutes,
	}
}

func switchStatus(status *bool) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 1,
		Value:      status,
	}
}

type FanLevel uint8

func (level FanLevel) Increase() FanLevel {
	if level == FanLevel4 {
		return FanLevel4
	}
	return FanLevel(int8(level) + 1)
}

func (level FanLevel) Decrease() FanLevel {
	if level == FanLevel1 {
		return FanLevel1
	}
	return FanLevel(uint8(level) - 1)
}

func (level FanLevel) Next() FanLevel {
	if nextLevel := level.Increase(); level != nextLevel {
		return nextLevel
	}
	return FanLevel1
}

const (
	FanLevel1 FanLevel = 1
	FanLevel2 FanLevel = 2
	FanLevel3 FanLevel = 3
	FanLevel4 FanLevel = 4
)

func fanLevel(level *FanLevel) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 2,
		Value:      level,
	}
}

type HorizontalAngle uint16

func (angle HorizontalAngle) Increase() HorizontalAngle {
	if angle == HorizontalAngle140 || angle == HorizontalAngle120 {
		return HorizontalAngle140
	}

	return HorizontalAngle(int16(angle) + 30)
}

func (angle HorizontalAngle) Decrease() HorizontalAngle {
	if angle == HorizontalAngle30 {
		return HorizontalAngle30
	} else if angle == HorizontalAngle140 {
		return HorizontalAngle120
	}
	return HorizontalAngle(uint16(angle) - 30)
}

func (angle HorizontalAngle) Next() HorizontalAngle {
	if nextAngle := angle.Increase(); angle != nextAngle {
		return nextAngle
	}
	return HorizontalAngle30
}

const (
	HorizontalAngle30 HorizontalAngle = 30
	HorizontalAngle60 HorizontalAngle = 60
	HorizontalAngle90 HorizontalAngle = 90
	HorizontalAngle120 HorizontalAngle = 120
	HorizontalAngle140 HorizontalAngle = 140
)

func horizontalAngle(angle *HorizontalAngle) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 5,
		Value:      angle,
	}
}

func mode(mode *Mode) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 3,
		Value:      mode,
	}
}