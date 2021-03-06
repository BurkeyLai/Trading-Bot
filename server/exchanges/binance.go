package exchanges

import (
	"context"
	"errors"
	"fmt"

	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

const (
	SPOT   = "spot"
	FUTURE = "future"
)

// BinanceWrapper represents the wrapper for the Binance exchange.
type BinanceWrapper struct {
	api              *binance.Client
	future           *futures.Client
	summaries        *SummaryCache
	candles          *CandlesCache
	orderbook        *OrderbookCache
	depositAddresses map[string]string
	websocketOn      bool
}

// NewBinanceWrapper creates a generic wrapper of the binance API.
func NewBinanceWrapper(spotPublicKey, spotSecretKey, futurePublicKey, futureSecretKey string, depositAddresses map[string]string) ExchangeWrapper {
	binance.UseTestnet = true
	futures.UseTestnet = true
	client := binance.NewClient(spotPublicKey, spotSecretKey)
	//future := futures.NewClient(publicKey, secretKey)
	future := binance.NewFuturesClient(futurePublicKey, futureSecretKey)
	return &BinanceWrapper{
		api:              client,
		future:           future,
		summaries:        NewSummaryCache(),
		candles:          NewCandlesCache(),
		orderbook:        NewOrderbookCache(),
		depositAddresses: depositAddresses,
		websocketOn:      false,
	}
}

// Name returns the name of the wrapped exchange.
func (wrapper *BinanceWrapper) Name() string {
	return "binance"
}

func (wrapper *BinanceWrapper) String() string {
	return wrapper.Name()
}

// GetMarkets Gets all the markets info.
func (wrapper *BinanceWrapper) GetMarkets(mode string) ([]*environment.Market, error) {
	//var symbolArray
	if mode == SPOT {
		binanceExchangeInfo, err := wrapper.api.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			return nil, err
		}
		ret := make([]*environment.Market, len(binanceExchangeInfo.Symbols))

		for i, market := range binanceExchangeInfo.Symbols {
			ret[i] = &environment.Market{
				Name:           market.Symbol,
				BaseCurrency:   market.BaseAsset,
				MarketCurrency: market.QuoteAsset,
				LotSizeMinQty:  market.LotSizeFilter().MinQuantity,
				LotSizeMaxQty:  market.LotSizeFilter().MaxQuantity,
				MinNotional:    market.MinNotionalFilter().MinNotional,
			}
		}
		return ret, nil
	} else {
		futureExchangeInfo, err := wrapper.future.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			return nil, err
		}
		ret := make([]*environment.Market, len(futureExchangeInfo.Symbols))

		for i, market := range futureExchangeInfo.Symbols {
			ret[i] = &environment.Market{
				Name:           market.Symbol,
				BaseCurrency:   market.BaseAsset,
				MarketCurrency: market.QuoteAsset,
				LotSizeMinQty:  market.LotSizeFilter().MinQuantity,
				LotSizeMaxQty:  market.LotSizeFilter().MaxQuantity,
				MinNotional:    market.MinNotionalFilter().Notional,
			}
		}
		return ret, nil
	}
}

// BuyLimit performs a limit buy action.
func (wrapper *BinanceWrapper) BuyLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeLimit).Side(binance.SideTypeBuy).Symbol(MarketNameFor(market, wrapper)).Price(fmt.Sprint(limit)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// SellLimit performs a limit sell action.
func (wrapper *BinanceWrapper) SellLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeLimit).Side(binance.SideTypeSell).Symbol(MarketNameFor(market, wrapper)).Price(fmt.Sprint(limit)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// BuyMarket performs a market buy action.
func (wrapper *BinanceWrapper) BuyMarket(market *environment.Market, amount float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeMarket).Side(binance.SideTypeBuy).Symbol(MarketNameFor(market, wrapper)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
	//return fmt.Sprint(orderNumber.OrderID), nil
}

// SellMarket performs a market sell action.
func (wrapper *BinanceWrapper) SellMarket(market *environment.Market, amount float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeMarket).Side(binance.SideTypeSell).Symbol(MarketNameFor(market, wrapper)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}

// CalculateTradingFees calculates the trading fees for an order on a specified market.
//
//     NOTE: In Binance fees are currently hardcoded.
func (wrapper *BinanceWrapper) CalculateTradingFees(market *environment.Market, amount float64, limit float64, orderType TradeType) float64 {
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
func (wrapper *BinanceWrapper) CalculateWithdrawFees(market *environment.Market, amount float64) float64 {
	panic("Not Implemented")
}

// FeedConnect connects to the feed of the exchange.
func (wrapper *BinanceWrapper) FeedConnect(markets []*environment.Market) error {
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
func (wrapper *BinanceWrapper) subscribeMarketSummaryFeed(market *environment.Market) error {
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

func (wrapper *BinanceWrapper) subscribeOrderbookFeed(market *environment.Market) {
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

func (wrapper *BinanceWrapper) orderbookFromREST(market *environment.Market) (*environment.OrderBook, int64, error) {
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
func (wrapper *BinanceWrapper) GetBalance(mode, symbol string) (*decimal.Decimal, error) {
	if mode == SPOT {
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
	} else {
		futureAccount, err := wrapper.future.NewGetAccountService().Do(context.Background())
		if err != nil {
			return nil, err
		}

		//for _, binanceBalance := range futureAccount.Balances {
		for _, asset := range futureAccount.Assets {
			if asset.Asset == symbol {
				ret, err := decimal.NewFromString(asset.WalletBalance)
				if err != nil {
					return nil, err
				}
				return &ret, nil
			}
		}

		return nil, errors.New("Symbol not found")
	}
}

// GetDepositAddress gets the deposit address for the specified coin on the exchange.
func (wrapper *BinanceWrapper) GetDepositAddress(coinTicker string) (string, bool) {
	addr, exists := wrapper.depositAddresses[coinTicker]
	return addr, exists
}

// GetMarketSummary gets the current market summary.
func (wrapper *BinanceWrapper) GetMarketSummary(market *environment.Market) (*environment.MarketSummary, error) {
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
func (wrapper *BinanceWrapper) GetCandles(market *environment.Market) ([]environment.CandleStick, error) {
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
func (wrapper *BinanceWrapper) GetOrderBook(market *environment.Market) (*environment.OrderBook, error) {
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
func (wrapper *BinanceWrapper) Withdraw(destinationAddress string, coinTicker string, amount float64) error {
	_, err := wrapper.api.NewCreateWithdrawService().Address(destinationAddress).Coin(coinTicker).Amount(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (wrapper *BinanceWrapper) AskOrderList(mode string, market *environment.Market) ([]*binance.Order, error) {
	if mode == SPOT {
		orders, err := wrapper.api.NewListOrdersService().Symbol(market.Name).Do(context.Background())
		if err != nil {
			fmt.Println(err)
		}
		//for _, o := range orders {
		//	fmt.Println(o)
		//}
		return orders, nil
	} else {

		return nil, nil
	}

}

func (wrapper *BinanceWrapper) QueryOrder(mode, id, symbol string) (*binance.Order, error) {
	if mode == SPOT {
		order, err := wrapper.api.NewGetOrderService().Symbol(symbol).
			OrigClientOrderID(id).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println(order)
		return order, nil
	} else {
		return nil, nil
	}
}
