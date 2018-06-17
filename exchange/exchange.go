package exchange

import (
	"fcoinExchange/conf"
	"fcoinExchange/fcoin"
	"fcoinExchange/log"
	"fcoinExchange/model"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

//
type Exchange struct {
	Symbol        string
	BaseCurrency  string
	QuoteCurrency string
	SellNumber    float64
	ExpectValue   float64
	Balance       map[string]*model.BalanceContext
	Quote         *model.Quote

	fcclient *fcoin.Client

	config *model.Configuration
	sync.RWMutex

	shuadanChan chan int
	accountChan chan int
}

//
func NewExchange(cfg *model.Configuration) (*Exchange, error) {
	client := fcoin.NewClient(cfg.AppKey, cfg.AppSecret, cfg.RequestTimeout)
	list, err := client.GetCurrencies()
	if err != nil {
		return nil, err
	}

	if list.Status != 0 {
		return nil, fmt.Errorf("get currencies but return status is not 0")
	}

	var (
		base  string
		quote string
	)
	for _, v := range list.Data {
		if strings.HasPrefix(cfg.Symbol, v) {
			base = v
			quote = cfg.Symbol[len(v):]
			break
		}
	}

	if !strings.Contains(strings.Join(list.Data, ","), quote) {
		return nil, fmt.Errorf("symbol %s does not support", cfg.Symbol)
	}

	return &Exchange{
		Symbol:        cfg.Symbol,
		BaseCurrency:  base,
		QuoteCurrency: quote,
		fcclient:      client,
		config:        cfg,
		Balance:       make(map[string]*model.BalanceContext),
		accountChan:   make(chan int, 1),
		shuadanChan:   make(chan int, 1),
	}, nil
}

func (p *Exchange) AutoUpdate() {
	//go p.AutoUpdateTicker()
	//go p.AutoUpdateBalance()
	if p.config.AutoCheckOrder {
		go p.AutoCheckOrders()
	}
	time.Sleep(time.Second)
	go p.AutoShuaDan()
}

// 自动更新行情信息
func (p *Exchange) AutoUpdateTicker() {
	log.Logger.Infof("start auto update ticker task")
	var (
		interval int64 = conf.GetConfiguration().UpdateTickerInterval
		ticker   *model.Ticker
		quote    *model.Quote
		err      error
	)

	if interval < 500 {
		log.Logger.Infof("update ticker interval less than 500, set it to 500")
		interval = 500
	}

	tk := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		ticker, err = p.fcclient.GetTicker(p.Symbol)
		if err != nil {
			log.Logger.Errorf("get %s ticker failed. %s\n", p.Symbol, err)
		} else {
			if ticker.Status != 0 {
				log.Logger.Errorf("get ticker but return status is %d", ticker.Status)
			} else {
				quote, err = fcoin.ParseTicker(ticker)
				if err != nil {
					log.Logger.Errorf("%s\n", err)
				} else {
					p.Quote = quote
				}

			}
		}
		<-tk.C
	}
}

func (p *Exchange) AutoCheckOrders() {
	log.Logger.Infof("start auto check order task")
	var (
		interval int64 = conf.GetConfiguration().CheckOrderInterval
		orders   *model.OrderList
		err      error
		querys   map[string]string = map[string]string{
			"symbol": p.Symbol,
			"limit":  "10",
		}
		states     []string = []string{"submitted", "partial_filled"}
		serverTime *model.ServerTime
		timeDValue int64
	)

	if interval < 500 {
		log.Logger.Infof("check order interval less than 500, set it to 500")
		interval = 500
	}

	tk := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		for _, state := range states {
			querys["states"] = state
			orders, err = p.fcclient.ListOrders(querys)
			if err != nil {
				log.Logger.Errorf("get orders failed. %s\n", err)
			} else {
				if orders.Status != 0 {
					log.Logger.Errorf("get orderlist but return status is %d", orders.Status)
				} else {
					serverTime, err = p.fcclient.GetServerTime()
					if err != nil {
						log.Logger.Errorf("get server time failed. %v", serverTime)
						serverTime = new(model.ServerTime)
						serverTime.Data = time.Now().UnixNano() / 1000000
					}
					for _, order := range orders.Data {
						log.Logger.Infof("order id %s, created at: %d", order.Id, order.CreatedAt)
						timeDValue = order.CreatedAt - serverTime.Data
						log.Logger.Infof("time d_value: %d", timeDValue)
						if math.Abs(float64(timeDValue)) > float64(p.config.RevokeOrderTime) {
							// invoke order
							go func(id string) {
								log.Logger.Infof("cancel order id %s", id)
								var corder *model.CancelOrder
								corder, err = p.fcclient.CancelOrder(order.Id)
								if err != nil {
									log.Logger.Infof("cancel order failed. %s", err)
									return
								}
								if corder.Status != 0 {
									log.Logger.Infof("cancel order failed. %v", corder)

								}
							}(order.Id)
						}
					}
				}
			}
		}
		<-tk.C
	}
}

func (p *Exchange) GetQuote() *model.Quote {
	p.RLock()
	defer p.RUnlock()
	return p.Quote
}

func (p *Exchange) Buy(price, amount string) (*model.Order, error) {
	return p.fcclient.CreateOrder(p.Symbol, "buy", "limit", price, amount)
}

func (p *Exchange) Sell(price, amount string) (*model.Order, error) {
	return p.fcclient.CreateOrder(p.Symbol, "sell", "limit", price, amount)
}

func (p *Exchange) GetAccountBalance() (*model.AccountBalance, error) {
	return p.fcclient.GetBalance()
}

func (p *Exchange) CancelOrders() {
	var (
		orders *model.OrderList
		err    error
		querys map[string]string = map[string]string{
			"symbol": p.Symbol,
			"limit":  "10",
		}
		states     []string = []string{"submitted", "partial_filled"}
		serverTime *model.ServerTime
		timeDValue int64
	)

	for _, state := range states {
		querys["states"] = state
		orders, err = p.fcclient.ListOrders(querys)
		if err != nil {
			log.Logger.Errorf("get orders failed. %s\n", err)
		} else {
			if orders.Status != 0 {
				log.Logger.Errorf("get orderlist but return status is %d", orders.Status)
			} else {
				serverTime, err = p.fcclient.GetServerTime()
				if err != nil {
					log.Logger.Errorf("get server time failed. %v", serverTime)
					serverTime = new(model.ServerTime)
					serverTime.Data = time.Now().UnixNano() / 1000000
				}
				for _, order := range orders.Data {
					log.Logger.Infof("order id %s, created at: %d", order.Id, order.CreatedAt)
					timeDValue = order.CreatedAt - serverTime.Data
					log.Logger.Infof("time d_value: %d", timeDValue)
					if math.Abs(float64(timeDValue)) > float64(p.config.RevokeOrderTime) {
						// invoke order
						log.Logger.Infof("cancel order id %s", order.Id)
						var corder *model.CancelOrder
						corder, err = p.fcclient.CancelOrder(order.Id)
						if err != nil {
							log.Logger.Infof("cancel order failed. %s", err)
						} else if corder.Status != 0 {
							log.Logger.Infof("cancel order failed. %v", corder)
						}
					}
					time.Sleep(time.Second)
				}
			}
		}
	}
}

func (p *Exchange) GetCurrentQuote() (*model.Quote, error) {
	ticker, err := p.fcclient.GetTicker(p.Symbol)
	if err != nil {
		return nil, err
	}

	if ticker.Status != 0 {
		return nil, err
	}

	return fcoin.ParseTicker(ticker)
}
