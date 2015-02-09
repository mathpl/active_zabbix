package active_zabbix

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type HostActiveKeys map[string]time.Duration

type ZabbixActiveClient struct {
	conn net.Conn
	addr *net.TCPAddr
	ZabbixActiveProto
}

func NewZabbixActiveClient(addr string, receive_timeout uint, send_timeout uint) (zc ZabbixActiveClient, err error) {
	addr = strings.Replace(addr, "zbx://", "", 1)
	zc.addr, err = net.ResolveTCPAddr("tcp", addr)
	zc.receive_timeout = time.Duration(receive_timeout) * time.Millisecond
	zc.send_timeout = time.Duration(send_timeout) * time.Millisecond
	return
}

func (zc *ZabbixActiveClient) getConn() (err error) {
	if zc.conn == nil {
		dialer := net.Dialer{}
		if zc.conn, err = dialer.Dial("tcp", zc.addr.String()); err != nil {
			return
		}
	}

	return
}

func (zc *ZabbixActiveClient) cleanupConn() {
	if zc.conn != nil {
		zc.conn.Close()
		zc.conn = nil
	}
}

func (zc *ZabbixActiveClient) ZabbixSendAndForget(data []byte) (err error) {
	err = zc.ZabbixSend(data)
	if err != nil {
		return
	}
	zc.cleanupConn()

	return
}

func (zc *ZabbixActiveClient) ZabbixSend(data []byte) (err error) {
	err = zc.getConn()
	if zc.conn == nil || err != nil {
		return
	}

	return zc.zabbixSend(zc.conn, data)
}

func (zc *ZabbixActiveClient) ZabbixReceive() (result []byte, err error) {
	err = zc.getConn()
	if zc.conn == nil || err != nil {
		return
	}
	defer zc.cleanupConn()

	return zc.zabbixReceive(zc.conn)
}

func (zc *ZabbixActiveClient) FetchActiveChecks(host string) (hc HostActiveKeys, err error) {
	msg := fmt.Sprintf("{\"request\":\"active checks\",\"host\":\"%s\"}", host)
	data := []byte(msg)

	hc = make(HostActiveKeys, 1)

	if err = zc.ZabbixSend(data); err != nil {
		return
	} else {
		var result []byte
		if result, err = zc.ZabbixReceive(); err != nil {
			return
		} else {
			// Parse json for key names
			var unmarshalledResult ActiveCheckResponseJson
			//Check what's the result on no keys

			err = json.Unmarshal(result, &unmarshalledResult)
			if err != nil {
				return
			}

			// Push key names for the current host
			for _, activeCheckKey := range unmarshalledResult.Data {
				if fDelay, ok := activeCheckKey.Delay.(float64); ok {
					hc[activeCheckKey.Key] = time.Duration(int(fDelay)) * time.Second
				} else if sDelay, ok := activeCheckKey.Delay.(string); ok {
					// Put 15 as delay if strconv doesn't work for now
					if delay, conv_err := strconv.ParseInt(sDelay, 10, 32); conv_err != nil {
						hc[activeCheckKey.Key] = time.Duration(15 * time.Second)
					} else {
						hc[activeCheckKey.Key] = time.Duration(int(delay)) * time.Second
					}
				}
			}
		}
	}

	return
}
