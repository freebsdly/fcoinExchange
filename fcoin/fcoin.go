package fcoin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fcoinExchange/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	baseUrl = "https://api.fcoin.com/v2"

	publicUrl  = fmt.Sprintf("%s/%s", baseUrl, "public")
	marketUrl  = fmt.Sprintf("%s/%s", baseUrl, "market")
	accountUrl = fmt.Sprintf("%s/%s", baseUrl, "accounts")
	orderUrl   = fmt.Sprintf("%s/%s", baseUrl, "orders")

	serverTimeUrl = fmt.Sprintf("%s/%s", publicUrl, "server-time")
	currencyUrl   = fmt.Sprintf("%s/%s", publicUrl, "currencies")
	symbolUrl     = fmt.Sprintf("%s/%s", publicUrl, "symbols")

	tickerUrl  = fmt.Sprintf("%s/%s", marketUrl, "ticker")
	balanceUrl = fmt.Sprintf("%s/%s", accountUrl, "balance")
)

type Client struct {
	client    *http.Client
	appKey    string
	appSecret []byte
	timeout   int
}

func NewClient(key, secret string, to int) *Client {
	var tp = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		client: &http.Client{
			Transport: tp,
			Timeout:   time.Duration(to) * time.Millisecond,
		},
		appKey:    key,
		appSecret: []byte(secret),
		timeout:   to,
	}
}

// signature msg
func (p *Client) Signature(msg string) string {
	bmsg := base64.StdEncoding.EncodeToString([]byte(msg))
	sh := hmac.New(sha1.New, p.appSecret)
	sh.Write([]byte(bmsg))
	return base64.StdEncoding.EncodeToString(sh.Sum(nil))
}

//
func (p *Client) MakeSignatureMessage(method, requrl string, timestamp int64, querys map[string]string) string {
	var (
		sigStr   string
		upMethod string = strings.ToUpper(method)
	)

	switch upMethod {
	case "GET":
		sigStr = fmt.Sprintf("%s%s%d", upMethod, requrl, timestamp)
	case "POST":
		sigStr = fmt.Sprintf("%s%s%d%s", upMethod, requrl, timestamp, SortMap(querys, "&"))
	}
	return sigStr
}

// get fcoin server time
func (p *Client) GetServerTime() (*model.ServerTime, error) {
	req, err := http.NewRequest("GET", serverTimeUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var st = new(model.ServerTime)

	err = json.Unmarshal(data, st)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// get coin types
func (p *Client) GetCurrencies() (*model.Currencies, error) {
	req, err := http.NewRequest("GET", currencyUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var crs = new(model.Currencies)

	err = json.Unmarshal(data, crs)
	if err != nil {
		return nil, err
	}

	return crs, nil
}

// get symbol type. example:
// ftusdt :  ft - usdt
// fteth  :  ft - eth
// ethusdt: eth - usdt
func (p *Client) GetSymbols() (*model.Symbols, error) {
	req, err := http.NewRequest("GET", symbolUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sbs = new(model.Symbols)

	err = json.Unmarshal(data, sbs)
	if err != nil {
		return nil, err
	}

	return sbs, nil
}

func (p *Client) GetTicker(symbol string) (*model.Ticker, error) {
	reqUrl := fmt.Sprintf("%s/%s", tickerUrl, symbol)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tk = new(model.Ticker)

	err = json.Unmarshal(data, tk)
	if err != nil {
		return nil, err
	}

	return tk, nil
}

func (p *Client) GetBalance() (*model.AccountBalance, error) {
	reqMethod := "GET"
	req, err := http.NewRequest(reqMethod, balanceUrl, nil)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixNano() / 1000000
	req.Header.Add("FC-ACCESS-KEY", p.appKey)
	req.Header.Add("FC-ACCESS-SIGNATURE", p.Signature(p.MakeSignatureMessage(reqMethod, balanceUrl, timestamp, nil)))
	req.Header.Add("FC-ACCESS-TIMESTAMP", fmt.Sprintf("%d", timestamp))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ab = new(model.AccountBalance)
	err = json.Unmarshal(data, ab)
	if err != nil {
		return nil, err
	}

	return ab, nil
}

func (p *Client) CreateOrder(symbol, side, otype, price, amount string) (*model.Order, error) {

	var (
		reqMethod = "POST"
		params    = make(map[string]string)
	)

	params["symbol"] = symbol
	params["side"] = side
	params["type"] = otype
	params["price"] = price
	params["amount"] = amount

	postBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(reqMethod, orderUrl, bytes.NewReader(postBody))
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixNano() / 1000000
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("FC-ACCESS-KEY", p.appKey)
	req.Header.Add("FC-ACCESS-SIGNATURE", p.Signature(p.MakeSignatureMessage(reqMethod, orderUrl, timestamp, params)))
	req.Header.Add("FC-ACCESS-TIMESTAMP", fmt.Sprintf("%d", timestamp))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var order = new(model.Order)
	err = json.Unmarshal(data, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (p *Client) CancelOrder(id string) (*model.CancelOrder, error) {

	var (
		reqMethod string = "POST"
		reqUrl    string
	)
	reqUrl = fmt.Sprintf("%s/%s/%s", orderUrl, id, "submit-cancel")
	req, err := http.NewRequest(reqMethod, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixNano() / 1000000
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("FC-ACCESS-KEY", p.appKey)
	req.Header.Add("FC-ACCESS-SIGNATURE", p.Signature(p.MakeSignatureMessage(reqMethod, reqUrl, timestamp, nil)))
	req.Header.Add("FC-ACCESS-TIMESTAMP", fmt.Sprintf("%d", timestamp))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var order = new(model.CancelOrder)
	err = json.Unmarshal(data, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (p *Client) ListOrders(querys map[string]string) (*model.OrderList, error) {

	var (
		reqMethod string = "GET"
		reqUrl    string
	)

	reqUrl = fmt.Sprintf("%s?%s", orderUrl, SortMap(querys, "&"))
	req, err := http.NewRequest(reqMethod, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixNano() / 1000000
	req.Header.Add("FC-ACCESS-KEY", p.appKey)
	req.Header.Add("FC-ACCESS-SIGNATURE", p.Signature(p.MakeSignatureMessage(reqMethod, reqUrl, timestamp, nil)))
	req.Header.Add("FC-ACCESS-TIMESTAMP", fmt.Sprintf("%d", timestamp))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orders = new(model.OrderList)
	err = json.Unmarshal(data, orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func ParseTicker(tk *model.Ticker) (*model.Quote, error) {
	if tk.Status != 0 {
		return nil, fmt.Errorf("status is not 0", tk.Status)
	}

	if len(tk.Data.Tickers) != 11 {
		return nil, fmt.Errorf("ticker format wrong. expected data ticker length is 11")
	}

	var qt = new(model.Quote)
	qt.Seq = tk.Data.Seq
	qt.Type = tk.Data.Type
	qt.LastestPrice = tk.Data.Tickers[0]
	qt.LastestVOL = tk.Data.Tickers[1]
	qt.MaxBuyOnePrice = tk.Data.Tickers[2]
	qt.MaxBuyNumber = tk.Data.Tickers[3]
	qt.MinSellOnePrice = tk.Data.Tickers[4]
	qt.MinSellNumber = tk.Data.Tickers[5]
	qt.TheDayBeforePrice = tk.Data.Tickers[6]
	qt.IntradayMaxPrice = tk.Data.Tickers[7]
	qt.IntradayMinPrice = tk.Data.Tickers[8]
	qt.IntradayBaseCurrencyVOL = tk.Data.Tickers[9]
	qt.IntradayQuoteCurrencyVOL = tk.Data.Tickers[10]

	return qt, nil

}
