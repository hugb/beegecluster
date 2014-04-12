package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	SUCCESS = "0"
	FAILURE = "1"
)

type Connection struct {
	Conn net.Conn
}

func PacketString(message string) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(len(message)))
	data = append(data, []byte(message)...)
	return data
}

func PacketByes(message []byte) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(len(message)))
	data = append(data, message...)
	return data
}

func (this *Connection) Read() (string, []byte, error) {
	head, err := this.packetData(2)
	if err != nil {
		return "", nil, err
	}
	length := int(binary.BigEndian.Uint16(head))
	data, err := this.packetData(length)
	if err != nil {
		return "", nil, err
	}
	cmd, payload := CmdDecode(length, data)
	return cmd, payload, nil
}

func (this *Connection) WriteBytes(code string, data []byte) error {
	_, err := this.Conn.Write(PacketByes(append(data, code...)))
	return err
}

func (this *Connection) WriteString(code string, data string) error {
	_, err := this.Conn.Write(PacketString(fmt.Sprintf("%s %s", data, code)))
	return err
}

func (this *Connection) WriteFails(data string) error {
	return this.WriteString(FAILURE, data)
}

func (this *Connection) WriteSuccess(data string) error {
	return this.WriteString(SUCCESS, data)
}

func (this *Connection) WriteSuccessBytes(data []byte) error {
	return this.WriteBytes(SUCCESS, data)
}

func (this *Connection) packetData(m int) (data []byte, err error) {
	data = make([]byte, m)
	for l, n := 0, 0; n < m; {
		l, err = this.Conn.Read(data[n:m])
		if nil != err && io.EOF != err {
			return data, err
		}
		n += l
	}
	return data, nil
}

func CmdDecode(length int, data []byte) (string, []byte) {
	blankIndex := length - 1
	for ; blankIndex > 0; blankIndex-- {
		if data[blankIndex] == 32 {
			break
		}
	}
	if blankIndex > 0 {
		return string(data[blankIndex+1 : length]), data[0:blankIndex]
	} else {
		return "", data
	}
}
