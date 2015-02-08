package active_zabbix

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type ActiveCheckKeyJson struct {
	Key string `json:"key"`

	//Zabbix 1.8 return these as string, 2+ as int.
	Delay       interface{} `json:"delay"`
	Lastlogsize interface{} `json:"lastlogsize"`
	Mtime       interface{} `json:"mtime"`
}

type ActiveCheckResponseJson struct {
	Response string               `json:"response"`
	Data     []ActiveCheckKeyJson `json:"data"`
}

type ZabbixMetricRequestJson struct {
	Request string                `json:"request"`
	Data    []ZabbixMetricKeyJson `json:"data"`
}

type ZabbixMetricKeyJson struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock string `json:"clock"`
}

type ZabbixActiveProto struct {
	receive_timeout time.Duration
	send_timeout    time.Duration
}

func (zp *ZabbixActiveProto) zabbixSend(conn net.Conn, data []byte) (err error) {
	zbxHeader := []byte("ZBXD\x01")
	// zabbix header + proto version + uint64 length
	zbxHeaderLength := len(zbxHeader) + 8

	dataLength := len(data)

	msgArray := make([]byte, zbxHeaderLength+dataLength)

	msgSlice := msgArray[0:0]
	msgSlice = append(msgSlice, zbxHeader...)

	byteBuff := make([]byte, 8)

	binary.LittleEndian.PutUint64(byteBuff, uint64(dataLength))
	msgSlice = append(msgSlice, byteBuff...)

	msgSlice = append(msgSlice, data...)

	conn.SetWriteDeadline(time.Now().Add(zp.send_timeout * time.Second))

	var n int
	if n, err = conn.Write(msgSlice); n != len(msgSlice) {
		err = fmt.Errorf("Full message not send, only %d of %d bytes", n, len(msgSlice))
	}

	return
}

func (zp *ZabbixActiveProto) zabbixReceive(conn net.Conn) (result []byte, err error) {
	// Get the response!
	conn.SetReadDeadline(time.Now().Add(zp.receive_timeout))

	// Fetch the header first to get the full length
	header := make([]byte, 13)
	//var header_length int
	var n int
	if n, err = io.ReadFull(conn, header); err != nil {
		return
	}

	// Check header content
	if string(header[:5]) != "ZBXD\x01" {
		err = fmt.Errorf("Unexpected response header from Zabbix: %s %d %s", string(header[:5]), len(header[:5]), header[:5])
		return
	}

	// Get length from zabbix protocol
	response_length := binary.LittleEndian.Uint64(header[5:13])

	conn.SetReadDeadline(time.Now().Add(zp.receive_timeout))

	// Get full reponse
	response := make([]byte, response_length)
	if n, err = io.ReadFull(conn, response); err != nil {
		return
	}

	if n != int(response_length) {
		err = fmt.Errorf("Unexpected response length from Zabbix header: %s", response_length)
		return
	}

	result = response

	return
}
