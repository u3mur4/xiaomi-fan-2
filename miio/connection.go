// Package miio implements the miIO binary protocol used by Xiaomi smart home
// devices. It provides an encrypted UDP connection over which JSON-RPC-style
// commands can be sent and received.
package miio

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"
)

type miio_connection struct {
	net.Conn
	reader      io.Reader
	packet      *packet
	iv          []byte
	key         []byte
	deviceID    uint32
	deviceToken []byte
	debug       io.Writer
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (miio *miio_connection) Read(b []byte) (n int, err error) {
	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.Magic)
	if err != nil {
		return 0, err
	}

	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.Length)
	if err != nil {
		return 0, err
	}

	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.Unknown1)
	if err != nil {
		return 0, err
	}

	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.DeviceID)
	if err != nil {
		return 0, err
	}

	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.Stamp)
	if err != nil {
		return 0, err
	}

	err = binary.Read(miio.reader, binary.BigEndian, &miio.packet.Checksum)
	if err != nil {
		return 0, err
	}

	length := miio.packet.Length - 0x20
	miio.packet.Data = make([]byte, length)
	err = binary.Read(miio.reader, binary.BigEndian, miio.packet.Data)
	if err != nil {
		return 0, err
	}

	if miio.debug != nil {
		miio.debugMsg("<-packet")
		data, err := miio.packet.serialize()
		if err != nil {
			miio.debugMsg(fmt.Sprintf("cannot serialize packet"))
		} else {
			miio.debugMsg(hex.Dump(data))
		}
	}

	decrypted, err := descryptPayload(miio.packet.Data, miio.key, miio.iv)
	if err != nil {
		return 0, err
	}

	if miio.debug != nil {
		miio.debugMsg("<-payload")
		miio.debugMsg(string(decrypted))
	}

	n = copy(b, decrypted)
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (miio *miio_connection) Write(b []byte) (n int, err error) {
	if miio.debug != nil {
		miio.debugMsg("->payload")
		miio.debugMsg(string(b))
	}

	encrypted, err := encryptPayload(b, miio.key, miio.iv)
	if err != nil {
		return 0, err
	}

	n, err = miio.packet.Write(encrypted)
	if err != nil {
		return 0, err
	}

	data, err := miio.packet.serialize()
	if err != nil {
		return 0, err
	}

	if miio.debug != nil {
		miio.debugMsg("->packet")
		miio.debugMsg(hex.Dump(data))
	}

	_, err = miio.Conn.Write(data)
	return
}

func (miio *miio_connection) debugMsg(msg string) {
	fmt.Fprintln(miio.debug, msg)
}

// sendHello performs the miIO handshake on an existing connection.
func sendHello(conn net.Conn) (token []byte, err error) {
	miio, ok := conn.(*miio_connection)
	if !ok {
		return nil, fmt.Errorf("not a miio connection")
	}

	hello := miio.packet
	hello.Length = 0x0020
	hello.Unknown1 = 0xFFFFFFFF
	hello.DeviceID = 0xFFFFFFFF
	hello.Stamp = 0xFFFFFFFF
	hello.Checksum = [16]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	hello.Data = []byte{}

	miio.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
	_, err = miio.Write(nil)
	if err != nil {
		return nil, err
	}

	miio.packet, _ = newPacket(miio.deviceID, miio.deviceToken)

	response := make([]byte, 1024)
	miio.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, err = miio.Read(response)
	if err != nil {
		return nil, err
	}
	return miio.packet.Checksum[:], nil
}

// Debug enables debug logging for a miIO connection. All sent and received
// packets (raw and decrypted) are written to the provided io.Writer.
func Debug(conn net.Conn, output io.Writer) (err error) {
	miio, ok := conn.(*miio_connection)
	if !ok {
		return fmt.Errorf("not a miio connection")
	}

	miio.debug = output
	return nil
}

// Dial opens an encrypted miIO connection to a Xiaomi device at the given
// IP address. It performs the initial handshake and returns a
// connection that transparently encrypts writes and decrypts reads.
func Dial(ip string, deviceID uint32, deviceToken []byte) (net.Conn, error) {

	connection, err := net.Dial("udp", ip)
	if err != nil {
		return nil, err
	}

	packet, err := newPacket(deviceID, deviceToken)
	if err != nil {
		return nil, err
	}

	key, iv, err := getKeyAndIV(deviceToken)
	if err != nil {
		return nil, err
	}

	miio := &miio_connection{
		Conn:        connection,
		reader:      bufio.NewReader(connection),
		packet:      packet,
		key:         key,
		iv:          iv,
		deviceID:    deviceID,
		deviceToken: deviceToken,
		debug:       nil,
	}

	_, err = sendHello(miio)
	if err != nil {
		return nil, err
	}

	return miio, nil
}
