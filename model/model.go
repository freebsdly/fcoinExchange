package model

const (
	TaskMode int = iota
	ScheduleMode
)

//
type Configuration struct {
	Mode                  int     `yaml:"mode"`
	ReloadInterval        int     `yaml:"reload_interval"`
	AppKey                string  `yaml:"appkey"`
	AppSecret             string  `yaml:"appsecret"`
	Symbol                string  `yaml:"symbol"`
	SellNumber            float64 `yaml:"sell_number"`
	MakeUpPercent         int     `yaml:"makeup_percent"`
	BalancePercent        int     `yaml:"balance_percent"`
	ExpectValue           float64 `yaml:"expect_value"`
	AutoCheckOrder        bool    `yaml:"auto_check_order"`
	CheckOrderInterval    int64   `yaml:"check_order_interval"`
	RevokeOrderTime       int64   `yaml:"revoke_order_time"`
	ShuaDanInterval       int64   `yaml:"shuadan_interval"`
	UpdateAccountInterval int64   `yaml:"update_account_interval"`
	UpdateTickerInterval  int64   `yaml:"update_ticker_interval"`
	RequestTimeout        int     `yaml:"request_timeout"`
	LogFile               string  `yaml:"log_file"`
	LogLevel              string  `yaml:"log_level"`
}

// 服务器时间
type ServerTime struct {
	Status int   `json:"status"`
	Data   int64 `json:"data"`
}

// 币种
type Currencies struct {
	Status int      `json:"status"`
	Data   []string `json:"data"`
}

// 交易対
type Symbols struct {
	Status int       `json:"status"`
	Data   []*Symbol `json:"data"`
}

type Symbol struct {
	Name          string `json:"name"`
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	PriceDecimal  int    `json:"price_decimal"`
	AmountDecimal int    `json:"amount_decimal"`
}

// 行情
type Ticker struct {
	Status int            `json:"status"`
	Data   *TickerContext `json:"data"`
}

type TickerContext struct {
	Type    string    `json:"type"`
	Seq     int64     `json:"seq"`
	Tickers []float64 `json:"ticker"`
}

// 行情报价
type Quote struct {
	Type                     string
	Seq                      int64
	LastestPrice             float64
	LastestVOL               float64
	MaxBuyOnePrice           float64
	MaxBuyNumber             float64
	MinSellOnePrice          float64
	MinSellNumber            float64
	TheDayBeforePrice        float64
	IntradayMaxPrice         float64
	IntradayMinPrice         float64
	IntradayBaseCurrencyVOL  float64
	IntradayQuoteCurrencyVOL float64
}

// 账户资产
type AccountBalance struct {
	Status int               `json:"status"`
	Data   []*BalanceContext `json:"data"`
}

type BalanceContext struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Frozen    string `json:"frozen"`
	Balance   string `json:"balance"`
}

//
type Order struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
}

//{
//  "status": 0,
//  "data": [
//    {
//      "id": "string",
//      "symbol": "string",
//      "type": "limit",
//      "side": "buy",
//      "price": "string",
//      "amount": "string",
//      "state": "submitted",
//      "executed_value": "string",
//      "fill_fees": "string",
//      "filled_amount": "string",
//      "created_at": 0,
//      "source": "web"
//    }
//  ]
//}

type OrderList struct {
	Status int          `json:"status"`
	Data   []*OrderInfo `json:"data"`
}

type OrderInfo struct {
	Id            string `json:"id"`
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	price         string `json:"price"`
	Amount        string `json:"amount"`
	State         string `json:"state"`
	ExecutedValue string `json:"executed_value"`
	FillFees      string `json:"fill_fees"`
	FilledAmount  string `json:"filled_amount"`
	CreatedAt     int64  `json:"created_at"`
	Source        string `json:"source"`
}

type CancelOrder struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   bool   `json:"data"`
}
