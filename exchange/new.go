package exchange

import (
	"fcoinExchange/conf"
	"fcoinExchange/log"
	"fcoinExchange/model"
	"fmt"
	"math"
	"strconv"
	"time"
)

//
func (p *Exchange) Start() {
	go p.AutoUpdateTicker()
	go p.AutoUpdateBalance()
	go p.AutoExchange()
	go p.AutoBalance()

}

func (p *Exchange) AutoUpdateBalance() {
	log.Logger.Infof("start auto update balance task")
	var (
		interval int64 = conf.GetConfiguration().UpdateAccountInterval
		balance  *model.AccountBalance
		err      error
	)

	if interval < 500 {
		log.Logger.Infof("update account interval less than 500, set it to 500")
		interval = 500
	}

	tk := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		balance, err = p.fcclient.GetBalance()
		if err != nil {
			log.Logger.Errorf("get balance failed. %s\n", err)
		} else {
			if balance.Status != 0 {
				log.Logger.Errorf("get balance but return status is %d", balance.Status)
			} else {
				for _, v := range balance.Data {
					p.Balance[v.Currency] = v
				}
			}
		}
		p.accountChan <- 1
		<-tk.C
	}
}

func (p *Exchange) AutoExchange() {
	log.Logger.Infof("start auto exchange")
	var (
		err    error
		quote  *model.Quote
		price  string
		number string
	)
	for {
		<-p.shuadanChan
		log.Logger.Infof("start exchange")

		quote, err = p.GetCurrentQuote()
		if err != nil {
			log.Logger.Errorf("get current quote failed. %s", err)
			continue
		}

		price = fmt.Sprintf("%.8f", math.Abs(quote.MaxBuyOnePrice+p.config.ExpectValue))
		number = fmt.Sprintf("%.2f", p.config.SellNumber)
		p.BuyAndSell(price, number)

	}
}

func (p *Exchange) AutoBalance() {
	log.Logger.Infof("start auto check balance")
	var (
		quote       *model.Quote
		err         error
		baseNumber  float64
		quoteNumber float64
	)

	for {
		<-p.accountChan

		baseNumber, err = strconv.ParseFloat(p.Balance[p.BaseCurrency].Available, 10)
		if err != nil {
			log.Logger.Errorf("%s", err)
			continue
		}
		quoteNumber, err = strconv.ParseFloat(p.Balance[p.QuoteCurrency].Available, 10)
		if err != nil {
			log.Logger.Errorf("%s", err)
			continue
		}

		// 获取行情
		quote, err = p.GetCurrentQuote()
		if err != nil {
			log.Logger.Errorf("get current quote failed. %s", err)
			continue
		}

		// 判断可用账户余额
		if quoteNumber < (quote.MinSellOnePrice*p.config.SellNumber) || baseNumber < p.config.SellNumber {
			log.Logger.Infof("account balance not enough, go to make up balance")
			p.MakeUpBalance()
			continue
		}

		p.shuadanChan <- 1

	}
}
