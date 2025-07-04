package fan2

func deviceManufacturer() *param {
	return &param{
		ServiceID:  1,
		PropertyID: 1,
	}
}

func deviceModel() *param {
	return &param{
		ServiceID:  1,
		PropertyID: 2,
	}
}

func deviceSerialNumber() *param {
	return &param{
		ServiceID:  1,
		PropertyID: 3,
	}
}

func deviceFirmwareVersion() *param {
	return &param{
		ServiceID:  1,
		PropertyID: 4,
	}
}
