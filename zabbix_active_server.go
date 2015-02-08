package active_zabbix

import (
	"encoding/json"
	"net"
	"time"
)

type ZabbixActiveServer struct {
	listener net.Listener
	addr     *net.TCPAddr
	ZabbixActiveProto
}

func NewZabbixActiveServer(addr string, receive_timeout uint, send_timeout uint) (zs ZabbixActiveServer, err error) {
	zs.receive_timeout = time.Duration(receive_timeout) * time.Millisecond
	zs.send_timeout = time.Duration(send_timeout) * time.Millisecond

	zs.addr, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	zs.listener, err = net.ListenTCP("tcp", zs.addr)
	if err != nil {
		return
	}

	return
}

func (zs *ZabbixActiveServer) Close() {
	zs.Close()
}

func (zs *ZabbixActiveServer) Listen(data_chan chan *ZabbixMetricRequestJson) error {
	for {
		conn, err := zs.listener.Accept()
		if err == nil {
			go zs.handle_connection(conn, data_chan)
		}
	}
}

func (zs *ZabbixActiveServer) handle_connection(conn net.Conn, data_chan chan *ZabbixMetricRequestJson) error {
	defer conn.Close()

	data, err := zs.zabbixReceive(conn)
	if err != nil {
		return err
	} else {
		var unmarshalledData ZabbixMetricRequestJson
		err = json.Unmarshal(data, &unmarshalledData)
		if err != nil {
			return err
		}
		data_chan <- &unmarshalledData
	}

	return err
}
