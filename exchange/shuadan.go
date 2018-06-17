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

func (p *Exchange) AutoShuaDan() {
	var (
		err         error
		quote       *model.Quote
		baseNumber  float64
		quoteNumber float64
		price       string
		number      string
		balance     *model.AccountBalance
		interval    = conf.GetConfiguration().ShuaDanInterval
	)
	if interval < 500 {
		interval = 500
	}

	tk := time.NewTicker(time.Duration(interval) * time.Millisecond)
	for {
		<-tk.C
		// 获取账户
		balance, err = p.GetAccountBalance()
		if err != nil {
			log.Logger.Errorf("get balance failed. %s\n", err)
			continue
		}

		if balance.Status != 0 {
			log.Logger.Errorf("get balance but return status is %d", balance.Status)
			continue
		}

		for _, v := range balance.Data {
			p.Balance[v.Currency] = v
		}

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
			log.Logger.Errorf("get quote failed. %s", err)
			continue
		}

		// 判断可用账户余额
		if quoteNumber < (quote.MinSellOnePrice*p.config.SellNumber) || baseNumber < p.config.SellNumber {
			log.Logger.Infof("account balance not enough, go to make up balance")
			p.MakeUpBalance()
			continue
		}

		price = fmt.Sprintf("%.8f", math.Abs(quote.MinSellOnePrice-p.config.ExpectValue))
		number = fmt.Sprintf("%.2f", p.config.SellNumber)
		p.BuyAndSell(price, number)

	}
}

func (p *Exchange) BuyAndSell(price, amount string) {
	var (
		order *model.Order
		err   error
	)
	go func() {
		order, err = p.Buy(price, amount)
		if err != nil || order.Status != 0 {
			log.Logger.Errorf("create buy order failed. %v. price: %s, amount: %s", order, price, amount)
		}
	}()

	go func() {
		order, err = p.Sell(price, amount)
		if err != nil || order.Status != 0 {
			log.Logger.Errorf("create sell order failed. %v, price: %s, amount: %s", order, price, amount)
		}
	}()
}

func (p *Exchange) SellAndBuy(price, amount string) {
	var (
		order *model.Order
		err   error
	)
	order, err = p.Sell(price, amount)
	if err != nil || order.Status != 0 {
		log.Logger.Errorf("create sell order failed. %v, price: %s, amount: %s", order, price, amount)
	}
	order, err = p.Buy(price, amount)
	if err != nil || order.Status != 0 {
		log.Logger.Errorf("create buy order failed. %v. price: %s, amount: %s", order, price, amount)
	}

}

func (p *Exchange) MakeUpBalance() error {
	var (
		err        error
		baseTotal  float64
		quoteTotal float64
		//quoteBaseTotal float64
		quote  *model.Quote
		bflag  int = 10
		qflag  int = 1
		code   int
		price  string
		amount string
	)

	// 获取账户余额
	baseTotal, err = strconv.ParseFloat(p.Balance[p.BaseCurrency].Balance, 10)
	if err != nil {
		log.Logger.Errorf("%s", err)
		return err
	}
	quoteTotal, err = strconv.ParseFloat(p.Balance[p.QuoteCurrency].Balance, 10)
	if err != nil {
		log.Logger.Errorf("%s", err)
		return err
	}

	quote, err = p.GetCurrentQuote()
	if err != nil {
		return err
	}

	if baseTotal > p.config.SellNumber {
		bflag = 20
	}

	if quoteTotal > quote.MaxBuyOnePrice*p.config.SellNumber {
		qflag = 2
	}

	code = bflag + qflag

	switch code {
	case 12:
		// 补充base currency
		log.Logger.Infof("make up %s currency", p.BaseCurrency)
		quote, err = p.GetCurrentQuote()
		if err != nil {
			return nil
		}
		price = fmt.Sprintf("%.8f", math.Abs(quote.MinSellOnePrice))
		amount = fmt.Sprintf("%.2f", p.config.SellNumber*float64(p.config.MakeUpPercent)/100)
		p.Buy(price, amount)
		break
	case 22:
		go p.CancelOrders()
		quote, err = p.GetCurrentQuote()
		if err != nil {
			return nil
		}
		price = fmt.Sprintf("%.8f", math.Abs(quote.MinSellOnePrice-p.config.ExpectValue))
		amount = fmt.Sprintf("%.2f", p.config.SellNumber*float64(p.config.BalancePercent)/100)
		p.BuyAndSell(price, amount)
		return nil
	case 21:
		// 补充quote currency
		log.Logger.Infof("make up %s currency", p.QuoteCurrency)
		quote, err = p.GetCurrentQuote()
		if err != nil {
			return nil
		}
		price = fmt.Sprintf("%.8f", math.Abs(quote.MaxBuyOnePrice))
		amount = fmt.Sprintf("%.2f", p.config.SellNumber*float64(p.config.MakeUpPercent)/100)
		p.Sell(price, amount)
		break
	case 11:
		// 减小sell number
		p.config.SellNumber = p.config.SellNumber * float64(p.config.BalancePercent/100)
		return nil
	}

	return nil
}
