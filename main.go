// fcoinExchange project main.go
package main

import (
	"fcoinExchange/conf"
	"fcoinExchange/exchange"
	"fcoinExchange/log"
)

func main() {
	conf.Init()
	log.Init()

	ex, err := exchange.NewExchange(conf.GetConfiguration())
	if err != nil {
		log.Logger.Fatalf("create exchange failed. %s\n", err)
	}

	switch conf.GetConfiguration().Mode {
	case 0:
		log.Logger.Infof("start task mode")
		ex.AutoUpdate()
	case 1:
		log.Logger.Infof("start schedule mode")
		ex.Start()
	default:
		log.Logger.Errorf("mode need be 0 or 1")
	}

	select {}
}
