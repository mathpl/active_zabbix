package main

import (
	"fmt"

	"github.com/mathpl/active_zabbix"
)

func main() {
	zs, err := active_zabbix.NewZabbixActiveServer("localhost:10052", 5000, 5000)

	if err != nil {
		fmt.Print(err)
	}

	metric_chan := make(chan *active_zabbix.ZabbixMetricRequestJson, 1)
	go zs.Listen(metric_chan)

	for metrics := range metric_chan {
		fmt.Printf("%+V\n", metrics)
		zs.Close()
	}
}
