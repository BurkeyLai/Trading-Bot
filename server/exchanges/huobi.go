package exchanges

import (
	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/huobirdcenter/huobi_golang/config"
	"github.com/huobirdcenter/huobi_golang/pkg/client"
)

// HuobiWrapper represents the wrapper for the Binance exchange.
type HuobiWrapper struct {
	apiCommon        *client.CommonClient
	summaries        *SummaryCache
	candles          *CandlesCache
	orderbook        *OrderbookCache
	depositAddresses map[string]string
	websocketOn      bool
}

// NewHuobiWrapper creates a generic wrapper of the Huobi API.
//func NewHuobiWrapper(publicKey string, secretKey string, depositAddresses map[string]string) ExchangeWrapper {
func NewHuobiWrapper(publicKey string, secretKey string, depositAddresses map[string]string) *HuobiWrapper {

	return &HuobiWrapper{
		apiCommon:        new(client.CommonClient).Init(config.Host),
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
func (wrapper *HuobiWrapper) GetMarkets() ([]*environment.Market, error) {

	//binanceSummary, err := wrapper.api.NewListPriceChangeStatsService().Symbol(MarketNameFor(market, wrapper)).Do(context.Background())
	//if err != nil {
	//	return nil, err
	//}
	// Get the timestamp from Huobi server and print on console

	//wrapper.apiCommon = new(client.CommonClient).Init(config.Host)
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

/*
// HuobiWrapper represents the wrapper for the Binance exchange.
type HuobiWrapper struct {
	api              *client
	summaries        *SummaryCache
	candles          *CandlesCache
	orderbook        *OrderbookCache
	depositAddresses map[string]string
	websocketOn      bool
}

// Name returns the name of the wrapped exchange.
func (wrapper *HuobiWrapper) Name() string {
	return "binance"
}

func (wrapper *HuobiWrapper) String() string {
	return wrapper.Name()
}

// BuyLimit performs a limit buy action.
func (wrapper *HuobiWrapper) BuyLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeLimit).Side(binance.SideTypeBuy).Symbol(MarketNameFor(market, wrapper)).Price(fmt.Sprint(limit)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// SellLimit performs a limit sell action.
func (wrapper *HuobiWrapper) SellLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeLimit).Side(binance.SideTypeSell).Symbol(MarketNameFor(market, wrapper)).Price(fmt.Sprint(limit)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// BuyMarket performs a market buy action.
func (wrapper *HuobiWrapper) BuyMarket(market *environment.Market, amount float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeMarket).Side(binance.SideTypeBuy).Symbol(MarketNameFor(market, wrapper)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
}

// SellMarket performs a market sell action.
func (wrapper *HuobiWrapper) SellMarket(market *environment.Market, amount float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeMarket).Side(binance.SideTypeSell).Symbol(MarketNameFor(market, wrapper)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// CalculateTradingFees calculates the trading fees for an order on a specified market.
//
//     NOTE: In Binance fees are currently hardcoded.
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

func (wrapper *HuobiWrapper) orderbookFromREST(market *environment.Market) (*environment.OrderBook, int64, error) {
	binanceOrderBook, err := wrapper.api.NewDepthService().Symbol(MarketNameFor(market, wrapper)).Do(context.Background())
	if err != nil {
		return nil, -1, err
	}

	var orderBook environment.OrderBook

	for _, ask := range binanceOrderBook.Asks {
		qty, err := decimal.NewFromString(ask.Quantity)
		if err != nil {
			return nil, -1, err
		}

		value, err := decimal.NewFromString(ask.Price)
		if err != nil {
			return nil, -1, err
		}

		orderBook.Asks = append(orderBook.Asks, environment.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	for _, bid := range binanceOrderBook.Bids {
		qty, err := decimal.NewFromString(bid.Quantity)
		if err != nil {
			return nil, -1, err
		}

		value, err := decimal.NewFromString(bid.Price)
		if err != nil {
			return nil, -1, err
		}

		orderBook.Bids = append(orderBook.Bids, environment.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	return &orderBook, binanceOrderBook.LastUpdateID, nil
}

// GetBalance gets the balance of the user of the specified currency.
func (wrapper *HuobiWrapper) GetBalance(symbol string) (*decimal.Decimal, error) {
	binanceAccount, err := wrapper.api.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	for _, binanceBalance := range binanceAccount.Balances {
		if binanceBalance.Asset == symbol {
			ret, err := decimal.NewFromString(binanceBalance.Free)
			if err != nil {
				return nil, err
			}
			return &ret, nil
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
		binanceSummary, err := wrapper.api.NewListPriceChangeStatsService().Symbol(MarketNameFor(market, wrapper)).Do(context.Background())
		if err != nil {
			return nil, err
		}

		ask, _ := decimal.NewFromString(binanceSummary[0].AskPrice)
		bid, _ := decimal.NewFromString(binanceSummary[0].BidPrice)
		high, _ := decimal.NewFromString(binanceSummary[0].HighPrice)
		low, _ := decimal.NewFromString(binanceSummary[0].LowPrice)
		volume, _ := decimal.NewFromString(binanceSummary[0].Volume)

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
func (wrapper *HuobiWrapper) GetCandles(market *environment.Market) ([]environment.CandleStick, error) {
	if !wrapper.websocketOn {
		binanceCandles, err := wrapper.api.NewKlinesService().Symbol(MarketNameFor(market, wrapper)).Do(context.Background())
		if err != nil {
			return nil, err
		}

		ret := make([]environment.CandleStick, len(binanceCandles))

		for i, binanceCandle := range binanceCandles {
			high, _ := decimal.NewFromString(binanceCandle.High)
			open, _ := decimal.NewFromString(binanceCandle.Open)
			close, _ := decimal.NewFromString(binanceCandle.Close)
			low, _ := decimal.NewFromString(binanceCandle.Low)
			volume, _ := decimal.NewFromString(binanceCandle.Volume)

			ret[i] = environment.CandleStick{
				High:   high,
				Open:   open,
				Close:  close,
				Low:    low,
				Volume: volume,
			}
		}

		wrapper.candles.Set(market, ret)
	}

	ret, candleLoaded := wrapper.candles.Get(market)
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
	_, err := wrapper.api.NewCreateWithdrawService().Address(destinationAddress).Coin(coinTicker).Amount(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}
*/
