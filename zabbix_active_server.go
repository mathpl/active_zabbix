package active_zabbix

import (
	"net"
	"time"
)

type ZabbixActiveServer struct {
	listener net.Listener
	addr     *net.TCPAddr
	ZabbixActiveProto
}

func NewZabbixServer(addr string, receive_timeout uint, send_timeout uint) (zs ZabbixActiveServer, err error) {
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
