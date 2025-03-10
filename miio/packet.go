package miio

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"math/rand"
)

// https://github.com/OpenMiHome/mihome-binary-protocol/blob/master/doc/PROTOCOL.md
type packet struct {
	Magic    uint16   // Magic number = 0x2131
	Length   uint16   // Packet Length (incl. header)
	Unknown1 uint32   // Unknown field.
	DeviceID uint32   // Device ID ("did")
	Stamp    uint32   // Stamp
	Checksum [16]byte // MD5 checksum or Device Token in response to the "Hello" packet
	Data     []byte   // optional variable-sized data (encrypted)

	reader io.Reader
	token  []byte
}

func (p *packet) Write(b []byte) (n int, err error) {
	if b == nil {
		return 0, nil
	}

	// save the encrypted payload
	p.Data = b[:]

	// recalculate packet length
	p.Length = uint16(0x20 + len(b))

	// increase stamp
	p.Stamp++

	// recalculate MD5 checksum
	err = p.calcMD5()
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (p *packet) calcMD5() error {
	// calculated for the whole packet including the MD5 field itself,
	// which must be initialized with the token value
	copy(p.Checksum[:], p.token)

	data, err := p.serialize()
	if err != nil {
		return err
	}

	// calculate checksum
	h := md5.New()
	_, err = h.Write(data)
	if err != nil {
		return err
	}

	// save result
	copy(p.Checksum[:], h.Sum(nil))
	return nil
}

func (p *packet) serialize() (data []byte, err error) {
	buf := bytes.NewBuffer(nil)
	buf.Grow(int(p.Length))

	err = binary.Write(buf, binary.BigEndian, &p.Magic)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, &p.Length)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, &p.Unknown1)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, &p.DeviceID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, &p.Stamp)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, p.Checksum)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, p.Data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func newPacket(deviceID uint32, token []byte) (*packet, error) {
	return &packet{
		Magic:    0x2131,
		Length:   0x20,
		Unknown1: 0,
		DeviceID: deviceID,
		Stamp:    rand.Uint32(),
		Checksum: [16]byte{},
		Data:     []byte{},

		token: token,
	}, nil

}
