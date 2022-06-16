package exchanges

import (
	"errors"
	"fmt"

	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/adshao/go-binance/v2"
	"github.com/huobirdcenter/huobi_golang/config"
	"github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
	"github.com/huobirdcenter/huobi_golang/pkg/model/order"
	"github.com/huobirdcenter/huobi_golang/pkg/model/wallet"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// HuobiWrapper represents the wrapper for the Huobi exchange.
type HuobiWrapper struct {
	apiCommon        *client.CommonClient
	apiOrder         *client.OrderClient
	apiMarket        *client.MarketClient
	apiAccount       *client.AccountClient
	apiWallet        *client.WalletClient
	accountId        string
	summaries        *SummaryCache
	candles          *CandlesCache
	orderbook        *OrderbookCache
	depositAddresses map[string]string
	websocketOn      bool
}

// NewHuobiWrapper creates a generic wrapper of the Huobi API.
func NewHuobiWrapper(spotPublicKey, spotSecretKey, futurePublicKey, futureSecretKey string, depositAddresses map[string]string) ExchangeWrapper {
	//func NewHuobiWrapper(publicKey string, secretKey string, depositAddresses map[string]string) *HuobiWrapper {
	api := new(client.AccountClient).Init(spotPublicKey, spotSecretKey, config.Host)
	resp, err := api.GetAccountInfo()
	var accountId string
	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		accountId = fmt.Sprint(resp[0].Id)
	}

	return &HuobiWrapper{
		apiCommon:        new(client.CommonClient).Init(config.Host), // config.Host = "api.huobi.pro" 為現貨網
		apiOrder:         new(client.OrderClient).Init(spotPublicKey, spotSecretKey, config.Host),
		apiMarket:        new(client.MarketClient).Init(config.Host),
		apiAccount:       new(client.AccountClient).Init(spotPublicKey, spotSecretKey, config.Host),
		apiWallet:        new(client.WalletClient).Init(spotPublicKey, spotSecretKey, config.Host),
		accountId:        accountId,
		summaries:        NewSummaryCache(),
		candles:          NewCandlesCache(),
		orderbook:        NewOrderbookCache(),
		depositAddresses: depositAddresses,
		websocketOn:      false,
	}
}

// Name returns the name of the wrapped exchange.
func (wrapper *HuobiWrapper) Name() string {
	return "huobi"
}

func (wrapper *HuobiWrapper) String() string {
	return wrapper.Name()
}

// GetMarkets Gets all the markets info.
func (wrapper *HuobiWrapper) GetMarkets(mode string) ([]*environment.Market, error) {
	api := wrapper.apiCommon
	//systemstatus, err := api.GetSystemStatus()
	//if err != nil {
	//	return nil, err
	//}
	//marketstatus, err := api.GetMarketStatus()
	//if err != nil {
	//	return nil, err
	//}
	symbols, err := api.GetSymbols()
	if err != nil {
		return nil, err
	}
	//currencys, err := api.GetCurrencys()
	//if err != nil {
	//	return nil, err
	//}
	////v2referencecurrencies, err := client.GetV2ReferenceCurrencies()
	//timestamp, err := api.GetTimestamp()
	//if err != nil {
	//	return nil, err
	//}

	ret := make([]*environment.Market, len(symbols))

	for i, market := range symbols {
		ret[i] = &environment.Market{
			Name:           market.Symbol,
			BaseCurrency:   market.BaseCurrency,
			MarketCurrency: market.QuoteCurrency,
		}
	}

	return ret, nil
}

// BuyLimit performs a limit buy action.
func (wrapper *HuobiWrapper) BuyLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	api := wrapper.apiOrder
	request := order.PlaceOrderRequest{
		AccountId: wrapper.accountId,
		Type:      "buy-limit",
		Source:    "spot-api",
		Symbol:    market.Name,
		Price:     fmt.Sprint(limit),
		Amount:    fmt.Sprint(amount),
	}
	resp, err := api.PlaceOrder(&request)
	if err != nil {
		return resp.ErrorCode + ": " + resp.ErrorMessage, err
	}
	return resp.Data, nil //The returned data object is a single string which represents the order id
}

// SellLimit performs a limit sell action.
func (wrapper *HuobiWrapper) SellLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	api := wrapper.apiOrder
	request := order.PlaceOrderRequest{
		AccountId: wrapper.accountId,
		Type:      "sell-limit",
		Source:    "spot-api",
		Symbol:    market.Name,
		Price:     fmt.Sprint(limit),
		Amount:    fmt.Sprint(amount),
	}
	resp, err := api.PlaceOrder(&request)
	if err != nil {
		return resp.ErrorCode + ": " + resp.ErrorMessage, err
	}
	return resp.Data, nil
}

// BuyMarket performs a market buy action.
func (wrapper *HuobiWrapper) BuyMarket(market *environment.Market, amount float64) (string, error) {
	api := wrapper.apiOrder
	request := order.PlaceOrderRequest{
		AccountId: wrapper.accountId,
		Type:      "buy-market",
		Source:    "spot-api",
		Symbol:    market.Name,
		Amount:    fmt.Sprint(amount),
	}
	resp, err := api.PlaceOrder(&request)
	if err != nil {
		return resp.ErrorCode + ": " + resp.ErrorMessage, err
	}
	return resp.Data, nil
}

// SellMarket performs a market sell action.
func (wrapper *HuobiWrapper) SellMarket(market *environment.Market, amount float64) (string, error) {
	api := wrapper.apiOrder
	request := order.PlaceOrderRequest{
		AccountId: wrapper.accountId,
		Type:      "sell-market",
		Source:    "spot-api",
		Symbol:    market.Name,
		Amount:    fmt.Sprint(amount),
	}
	resp, err := api.PlaceOrder(&request)
	if err != nil {
		return resp.ErrorCode + ": " + resp.ErrorMessage, err
	}
	return resp.Data, nil
}

// CalculateTradingFees calculates the trading fees for an order on a specified market.
//
//     NOTE: In Huobi fees are currently hardcoded.
func (wrapper *HuobiWrapper) CalculateTradingFees(market *environment.Market, amount float64, limit float64, orderType TradeType) float64 {
	var feePercentage float64
	if orderType == MakerTrade {
		feePercentage = 0.0010
	} else if orderType == TakerTrade {
		feePercentage = 0.0010
	} else {
		panic("Unknown trade type")
	}

	return amount * limit * feePercentage
}

// CalculateWithdrawFees calculates the withdrawal fees on a specified market.
func (wrapper *HuobiWrapper) CalculateWithdrawFees(market *environment.Market, amount float64) float64 {
	panic("Not Implemented")
}

// FeedConnect connects to the feed of the exchange.
func (wrapper *HuobiWrapper) FeedConnect(markets []*environment.Market) error {
	for _, m := range markets {
		err := wrapper.subscribeMarketSummaryFeed(m)
		if err != nil {
			return err
		}
		wrapper.subscribeOrderbookFeed(m)
	}
	wrapper.websocketOn = true

	return nil
}

// SubscribeMarketSummaryFeed subscribes to the Market Summary Feed service.
func (wrapper *HuobiWrapper) subscribeMarketSummaryFeed(market *environment.Market) error {
	//market.$symbol.ticker
	_, _, err := binance.WsMarketStatServe(MarketNameFor(market, wrapper), func(event *binance.WsMarketStatEvent) {
		high, _ := decimal.NewFromString(event.HighPrice)
		low, _ := decimal.NewFromString(event.LowPrice)
		ask, _ := decimal.NewFromString(event.AskPrice)
		bid, _ := decimal.NewFromString(event.BidPrice)
		last, _ := decimal.NewFromString(event.LastPrice)
		volume, _ := decimal.NewFromString(event.BaseVolume)

		wrapper.summaries.Set(market, &environment.MarketSummary{
			High:   high,
			Low:    low,
			Ask:    ask,
			Bid:    bid,
			Last:   last,
			Volume: volume,
		})
	}, func(error) {})
	if err != nil {
		return err
	}
	return nil
}

func (wrapper *HuobiWrapper) subscribeOrderbookFeed(market *environment.Market) {
	go func() {
		for {
			_, lastUpdateID, err := wrapper.orderbookFromREST(market)
			if err != nil {
				logrus.Error(err)
				return
			}
			// 24 hours max
			currentUpdateID := lastUpdateID

			done, _, _ := binance.WsPartialDepthServe(MarketNameFor(market, wrapper), "20", func(event *binance.WsPartialDepthEvent) {
				if event.LastUpdateID <= currentUpdateID { // this update is more recent than the latest fetched
					return
				}

				var orderbook environment.OrderBook

				orderbook.Asks = make([]environment.Order, len(event.Asks))
				orderbook.Bids = make([]environment.Order, len(event.Bids))

				for i, ask := range event.Asks {
					price, _ := decimal.NewFromString(ask.Price)
					quantity, _ := decimal.NewFromString(ask.Quantity)
					newOrder := environment.Order{
						Value:    price,
						Quantity: quantity,
					}
					orderbook.Asks[i] = newOrder
				}

				for i, bid := range event.Bids {
					price, _ := decimal.NewFromString(bid.Price)
					quantity, _ := decimal.NewFromString(bid.Quantity)
					newOrder := environment.Order{
						Value:    price,
						Quantity: quantity,
					}
					orderbook.Bids[i] = newOrder
				}

				wrapper.orderbook.Set(market, &orderbook)
			}, func(err error) {
				logrus.Error(err)
			})

			<-done
		}
	}()
}

func (wrapper *HuobiWrapper) orderbookFromREST(m *environment.Market) (*environment.OrderBook, int64, error) {
	api := wrapper.apiMarket
	optionalRequest := market.GetDepthOptionalRequest{Size: 10}
	resp, err := api.GetDepth(m.Name, market.STEP0, optionalRequest)
	if err != nil {
		return nil, -1, err
	}

	var orderBook environment.OrderBook

	for _, ask := range resp.Asks {
		value, err := decimal.NewFromString(ask[0].String())
		if err != nil {
			return nil, -1, err
		}

		qty, err := decimal.NewFromString(ask[1].String())
		if err != nil {
			return nil, -1, err
		}

		orderBook.Asks = append(orderBook.Asks, environment.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	for _, bid := range resp.Bids {
		value, err := decimal.NewFromString(bid[0].String())
		if err != nil {
			return nil, -1, err
		}

		qty, err := decimal.NewFromString(bid[1].String())
		if err != nil {
			return nil, -1, err
		}

		orderBook.Bids = append(orderBook.Bids, environment.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	//return &orderBook, resp.Version, nil
	return &orderBook, resp.Timestamp, nil
}

// GetBalance gets the balance of the user of the specified currency.
func (wrapper *HuobiWrapper) GetBalance(mode, symbol string) (*decimal.Decimal, error) {

	api := wrapper.apiAccount
	resp, err := api.GetAccountBalance(wrapper.accountId)
	if err != nil {
		return nil, errors.New("Get account balance error: " + err.Error())
	} else {
		fmt.Printf("Get account balance: id=%d, type=%s, state=%s, count=%d",
			resp.Id, resp.Type, resp.State, len(resp.List))
		if resp.List != nil {
			for _, result := range resp.List {
				if result.Currency == symbol {
					ret, err := decimal.NewFromString(result.Balance)
					if err != nil {
						return nil, err
					}
					return &ret, nil
				}
			}
		}
	}

	return nil, errors.New("Symbol not found")

}

// GetDepositAddress gets the deposit address for the specified coin on the exchange.
func (wrapper *HuobiWrapper) GetDepositAddress(coinTicker string) (string, bool) {
	addr, exists := wrapper.depositAddresses[coinTicker]
	return addr, exists
}

// GetMarketSummary gets the current market summary.
func (wrapper *HuobiWrapper) GetMarketSummary(market *environment.Market) (*environment.MarketSummary, error) {
	if !wrapper.websocketOn {
		api := wrapper.apiMarket
		resp, err := api.GetLast24hCandlestickAskBid(market.Name)
		if err != nil {
			return nil, err
		}

		ask, _ := decimal.NewFromString(resp.Ask[0].String())
		bid, _ := decimal.NewFromString(resp.Bid[0].String())
		high, _ := decimal.NewFromString(resp.High.String())
		low, _ := decimal.NewFromString(resp.Low.String())
		volume, _ := decimal.NewFromString(resp.Vol.String())

		wrapper.summaries.Set(market, &environment.MarketSummary{
			Last:   ask,
			Ask:    ask,
			Bid:    bid,
			High:   high,
			Low:    low,
			Volume: volume,
		})
	}

	ret, summaryLoaded := wrapper.summaries.Get(market)
	if !summaryLoaded {
		return nil, errors.New("Summary not loaded")
	}

	return ret, nil
}

// GetCandles gets the candle data from the exchange.
func (wrapper *HuobiWrapper) GetCandles(m *environment.Market) ([]environment.CandleStick, error) {
	if !wrapper.websocketOn {
		api := wrapper.apiMarket
		optionalRequest := market.GetCandlestickOptionalRequest{Period: market.MIN1, Size: 10}
		resp, err := api.GetCandlestick(m.Name, optionalRequest)
		if err != nil {
			return nil, err
		}

		ret := make([]environment.CandleStick, len(resp))

		for i, kline := range resp {
			high, _ := decimal.NewFromString(kline.High.String())
			open, _ := decimal.NewFromString(kline.Open.String())
			close, _ := decimal.NewFromString(kline.Close.String())
			low, _ := decimal.NewFromString(kline.Low.String())
			volume, _ := decimal.NewFromString(kline.Vol.String())

			ret[i] = environment.CandleStick{
				High:   high,
				Open:   open,
				Close:  close,
				Low:    low,
				Volume: volume,
			}
		}

		wrapper.candles.Set(m, ret)
	}

	ret, candleLoaded := wrapper.candles.Get(m)
	if !candleLoaded {
		return nil, errors.New("No candle data yet")
	}

	return ret, nil
}

// GetOrderBook gets the order(ASK + BID) book of a market.
func (wrapper *HuobiWrapper) GetOrderBook(market *environment.Market) (*environment.OrderBook, error) {
	if !wrapper.websocketOn {
		orderbook, _, err := wrapper.orderbookFromREST(market)
		if err != nil {
			return nil, err
		}

		wrapper.orderbook.Set(market, orderbook)
		return orderbook, nil
	}

	orderbook, exists := wrapper.orderbook.Get(market)
	if !exists {
		return nil, errors.New("Orderbook not loaded")
	}

	return orderbook, nil
}

// Withdraw performs a withdraw operation from the exchange to a destination address.
func (wrapper *HuobiWrapper) Withdraw(destinationAddress string, coinTicker string, amount float64) error {

	//_, err := wrapper.api.NewCreateWithdrawService().Address(destinationAddress).Coin(coinTicker).Amount(fmt.Sprint(amount)).Do(context.Background())
	api := wrapper.apiWallet
	createWithdrawRequest := wallet.CreateWithdrawRequest{
		Address:  destinationAddress,
		Amount:   fmt.Sprint(amount),
		Currency: "usdt",
		Fee:      "1.0"} //CalculateWithdrawFees
	resp, err := api.CreateWithdraw(createWithdrawRequest)
	if err != nil {
		return err
	}
	fmt.Printf("Create withdraw request successfully: id=%d", resp)

	return nil
}

func (wrapper *HuobiWrapper) AskOrderList(mode string, market *environment.Market) ([]*binance.Order, error) {
	if mode == SPOT {

		return nil, nil
	} else {

		return nil, nil
	}

}
