package main

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
	"github.com/saniales/golang-crypto-trading-bot/environment"
)

// BinanceWrapper represents the wrapper for the Binance exchange.
type BinanceWrapper struct {
	api              *binance.Client
	summaries        *SummaryCache
	candles          *CandlesCache
	orderbook        *OrderbookCache
	depositAddresses map[string]string
	websocketOn      bool
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
}

// SellMarket performs a market sell action.
func (wrapper *BinanceWrapper) SellMarket(market *environment.Market, amount float64) (string, error) {
	orderNumber, err := wrapper.api.NewCreateOrderService().Type(binance.OrderTypeMarket).Side(binance.SideTypeSell).Symbol(MarketNameFor(market, wrapper)).Quantity(fmt.Sprint(amount)).Do(context.Background())
	if err != nil {
		return "", err
	}
	return orderNumber.ClientOrderID, nil
}
