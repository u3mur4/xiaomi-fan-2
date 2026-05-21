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
	connection  net.Conn
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

	_, err = miio.connection.Write(data)
	return
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (miio *miio_connection) Close() error {
	return miio.connection.Close()
}

// LocalAddr returns the local network address.
func (miio *miio_connection) LocalAddr() net.Addr {
	return miio.connection.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (miio *miio_connection) RemoteAddr() net.Addr {
	return miio.connection.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (miio *miio_connection) SetDeadline(t time.Time) error {
	return miio.connection.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (miio *miio_connection) SetReadDeadline(t time.Time) error {
	return miio.connection.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (miio *miio_connection) SetWriteDeadline(t time.Time) error {
	return miio.connection.SetWriteDeadline(t)
}

func (miio *miio_connection) debugMsg(msg string) {
	fmt.Fprintln(miio.debug, msg)
}

func SendHello(conn net.Conn) (token []byte, err error) {
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

func Debug(conn net.Conn, output io.Writer) (err error) {
	miio, ok := conn.(*miio_connection)
	if !ok {
		return fmt.Errorf("not a miio connection")
	}

	miio.debug = output
	return nil
}

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
		connection:  connection,
		reader:      bufio.NewReader(connection),
		packet:      packet,
		key:         key,
		iv:          iv,
		deviceID:    deviceID,
		deviceToken: deviceToken,
		debug:       nil,
	}

	_, err = SendHello(miio)
	if err != nil {
		return nil, err
	}

	return miio, nil
}
