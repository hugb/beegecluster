package utils

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	SUCCESS = "0"
	FAILURE = "1"
)

type Connection struct {
	Src  string
	Conn net.Conn
}

//封包，在真实数据前增加两个字节的数据长度
func PacketString(message string) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(len(message)))
	data = append(data, []byte(message)...)
	return data
}

// 功能同上
func PacketByes(message []byte) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(len(message)))
	data = append(data, message...)
	return data
}

func (this *Connection) Read() (int, []byte, error) {
	// 读取包的头部，头部为两个字节的包长度，使用此封包结构目的是防止tcp粘包
	head, err := this.packetData(2)
	if err != nil {
		return 0, nil, err
	}
	length := int(binary.BigEndian.Uint16(head))
	// 读取真实数据
	data, err := this.packetData(length)
	if err != nil {
		return 0, nil, err
	}
	return length, data, nil
}

//封装成固定的消息格式"data flag"，数据与表示以空格分隔
func (this *Connection) SendCommandString(cmd string, data string) (int, error) {
	return this.Conn.Write(PacketString(fmt.Sprintf("%s %s", data, cmd)))
}

func (this *Connection) SendCommandBytes(cmd string, data []byte) (int, error) {
	data = append(data, 32)
	data = append(data, cmd...)
	return this.Conn.Write(PacketByes(data))
}

// 读取m个字节的数据
func (this *Connection) packetData(m int) (data []byte, err error) {
	data = make([]byte, m)
	for l, n := 0, 0; n < m; {
		l, err = this.Conn.Read(data[n:m])
		if nil != err {
			return data, err
		}
		n += l
	}
	return data, nil
}

//解析命令，从"payload command"得到命令为command，数据为payload
func CmdDecode(length int, data []byte) (string, []byte) {
	// 找到分隔符位置
	blankIndex := length - 1
	for ; blankIndex >= 0; blankIndex-- {
		if data[blankIndex] == 32 {
			break
		}
	}
	// 解析出命令和数据
	if blankIndex >= 0 {
		return string(data[blankIndex+1 : length]), data[0:blankIndex]
	} else {
		return "", data
	}
}
