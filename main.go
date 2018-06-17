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
	ex.AutoUpdate()
	//ex.Start()

	select {}
}
