package fan1c

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
		PropertyID: 3,
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
	if level == FanLevel1 {
		return FanLevel2
	}
	return FanLevel3
}

func (level FanLevel) Decrease() FanLevel {
	if level == FanLevel3 {
		return FanLevel2
	}
	return FanLevel1
}

func (level FanLevel) Next() FanLevel {
	if level == FanLevel1 {
		return FanLevel2
	} else if level == FanLevel2 {
		return FanLevel3
	}
	return FanLevel1
}

const (
	FanLevel1 FanLevel = 1
	FanLevel2 FanLevel = 2
	FanLevel3 FanLevel = 3
)

func fanLevel(level *FanLevel) *param {
	return &param{
		ServiceID:  2,
		PropertyID: 2,
		Value:      level,
	}
}
