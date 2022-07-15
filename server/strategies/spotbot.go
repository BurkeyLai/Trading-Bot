// Copyright © 2017 Alessandro Sanino <saninoale@gmail.com>
//
// Ths program s free software: you can redstribute it and/or modify
// it under the terms of the GNU General Public License as publshed by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Ths program s dstributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ths program. If not, see <http://www.gnu.org/licenses/>.

package strategies

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/BurkeyLai/Trading-Bot/server/exchanges"
	"github.com/BurkeyLai/Trading-Bot/server/proto"
	"github.com/BurkeyLai/Trading-Bot/server/utils"
	"google.golang.org/api/iterator"
)

// SpotBotStrategy is an interval based strategy.
type SpotBotStrategy struct {
	UserName       string
	UserId         string
	CycleType      string
	Model          StrategyModel
	OpenPrice      float64
	AvgPrice       float64
	Qty            float64
	ActivePercent  float64
	ReversePercent float64
	OrderIdList    []string
	Interval       time.Duration
	Stream         proto.TradingBot_CreateStreamServer
	Market         *environment.Market
	Doc            *firestore.DocumentRef
	Online         bool
	ClosePosition  chan bool
	ShutDown       chan bool
}

// Name returns the name of the strategy.
func (s SpotBotStrategy) Name() string {
	return s.Model.Name
}

// String returns a string representation of the object.
func (s SpotBotStrategy) String() string {
	return s.Name()
}

func (s SpotBotStrategy) UpdateAvgPrice(wrapper exchanges.ExchangeWrapper, market *environment.Market) float64 {
	TotalUsd := 0.0
	TotalQty := 0.0

	orders, err := wrapper.AskOrderList("spot", market)
	if err != nil {
		fmt.Println(err)
	}
	market.Orders = orders

	for _, orderId := range s.OrderIdList {
		for _, order := range market.Orders {
			if orderId == order.ClientOrderID {
				usd, _ := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)
				qty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
				if order.Side == "BUY" {
					TotalUsd += usd
					TotalQty += qty
				}
				break
			}
		}
	}
	fmt.Println("TotalUsd: " + fmt.Sprint(TotalUsd))
	fmt.Println("TotalQty: " + fmt.Sprint(TotalQty))
	if TotalUsd >= 0 && TotalQty > 0 {
		AvgPrice := TotalUsd / TotalQty
		return AvgPrice
	} else {
		fmt.Println("幣種存額小於等於0，賣出的幣種數量已高於此機器人所購買的幣種的量")
		return 0.0
	}
}

func (s SpotBotStrategy) UpdateBotInfo(content, exch, balance string) {
	if s.Online {
		msg := &proto.Message{
			User: &proto.User{
				Id:   s.UserId,
				Name: s.UserName,
			},
			Botinfo: &proto.BotInfo{
				Exch:          exch,
				Mode:          "spot",
				Modelname:     s.Model.Name,
				Avgprice:      fmt.Sprint(s.AvgPrice),
				Symbolbalance: balance,
				Orderidlist:   s.OrderIdList,
			},
			Content:   content,
			Timestamp: time.Now().String(),
		}
		s.Stream.Send(msg)
	} else {
		if s.Model.Name != "" {
			ctx := context.Background()
			userSnap := s.Doc.Snapshots(ctx)

			for {
				snap, err := userSnap.Next()
				if err == iterator.Done {
					break
				}
				bots_array, err := snap.DataAt("bots_array")
				if bots_array != nil {
					switch t := bots_array.(type) {
					case []interface{}:
						var botExchname, botMode, botSymbol string
						for _, bot := range t {
							// Search the bot data
							val := reflect.ValueOf(bot)
							if val.Kind() == reflect.Map {
								for _, key := range val.MapKeys() {
									v := val.MapIndex(key)
									switch value := v.Interface().(type) {
									case string:
										switch key.String() {
										case "symbol":
											botSymbol = value
										case "exchange":
											botExchname = value
										case "mode":
											botMode = value
										}
									}
								}
								if botExchname == exch && botMode == "spot" && botSymbol == s.Model.Name {
									// Update bot data in Firestore
									val.SetMapIndex(reflect.ValueOf("average_price"), reflect.ValueOf(fmt.Sprint(s.AvgPrice)))
									val.SetMapIndex(reflect.ValueOf("order_id_list"), reflect.ValueOf(s.OrderIdList))
									val.SetMapIndex(reflect.ValueOf("symbol_balance"), reflect.ValueOf(balance))

									s.Doc.Set(ctx, map[string]interface{}{
										"bots_array": bots_array,
									}, firestore.MergeAll)
									break
								}
							}
						}

					}
				}
				break
			}
		}
	}
}

func (s SpotBotStrategy) LaunchOrder(wrapper exchanges.ExchangeWrapper, market *environment.Market, price float64, try_again *bool, action string) string {
	lotSizeMinQty, _ := strconv.ParseFloat(market.LotSizeMinQty, 64)
	minNotional, _ := strconv.ParseFloat(market.MinNotional, 64)

	var clientOrderId string
	var err error
	if action == "buy" {
		amount, err := utils.CorrectPrecision(s.Qty/price, price, lotSizeMinQty, minNotional)
		if err != nil {
			return ""
		}
		fmt.Println("quantity: " + fmt.Sprint(s.Qty))
		fmt.Println("lastPrice: " + fmt.Sprint(price))
		fmt.Println("amount: " + fmt.Sprint(amount))
		requestSymbol := market.MarketCurrency
		if wrapper.Name() == "binance" {
			requestSymbol = strings.ToUpper(requestSymbol)
		}
		b, err := wrapper.GetBalance("spot", requestSymbol)
		if err != nil {
			return ""
		}
		balance, _ := strconv.ParseFloat(b.String(), 64)
		if balance >= s.Qty {
			clientOrderId, err = wrapper.BuyMarket(market, amount)
		} else {
			return ""
		}
	} else if action == "sell" {
		TotalQty := 0.0
		orders, err := wrapper.AskOrderList("spot", market)
		if err != nil {
			fmt.Println(err)
		}
		market.Orders = orders
		for _, orderId := range s.OrderIdList {
			for _, order := range market.Orders {
				if orderId == order.ClientOrderID {
					qty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
					if order.Side == "BUY" {
						TotalQty += qty
					}
					break
				}
			}
		}

		amount, err := utils.CorrectPrecision(TotalQty, price, lotSizeMinQty, minNotional)
		if err != nil {
			return ""
		}

		balance, _ := strconv.ParseFloat(market.Balance, 64)
		if balance >= TotalQty {
			clientOrderId, err = wrapper.SellMarket(market, amount)
		} else {
			return ""
		}
	}
	if err != nil {
		fmt.Println(err)
		*try_again = true
		return ""
	}
	//clientOrderId, _ := wrapper.SellMarket(m, amount)
	//resp.Orderid = clientOrderId

	fmt.Println("clientOrderId: === " + clientOrderId + " ===")
	*try_again = false
	return clientOrderId
}

func (s SpotBotStrategy) CalculateProfit(open, close float64, orderid string, wrapper exchanges.ExchangeWrapper) (float64, error) {
	order, err := wrapper.QueryOrder("spot", orderid, s.Model.Name)
	if err != nil {
		return 0.0, err
	}

	if order == nil {
		return 0, nil
	}
	amount, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	profit := (close - open) * amount
	return profit, nil
}

// Apply executes Cyclically the On Update, basing on provided interval.
func (s SpotBotStrategy) Apply(wrappers []exchanges.ExchangeWrapper, markets []*environment.Market) {
	var err error
	var last_lastPrice_low, last_lastPrice_high float64
	var buy_again, sell_again bool
	var isShutDown bool
	var ok bool
	buy_again = false
	sell_again = false
	isShutDown = false
	ok = true

	s.ClosePosition <- false
	s.ShutDown <- false
	isShutDown, ok = <-s.ShutDown

	hasSetupFunc := s.Model.Setup != nil
	hasTearDownFunc := s.Model.TearDown != nil
	hasUpdateFunc := s.Model.OnUpdate != nil
	hasErrorFunc := s.Model.OnError != nil

	if hasSetupFunc {
		err = s.Model.Setup(wrappers, markets)
		if err != nil && hasErrorFunc {
			s.Model.OnError(err)
		}

		if s.Online {
			s.AvgPrice, _ = markets[0].Summary.Last.Float64()
		} else {
			s.AvgPrice = s.UpdateAvgPrice(wrappers[0], markets[0])
			if s.AvgPrice == 0.0 {
				isShutDown = true
			}
		}
		s.Market = markets[0]
		last_lastPrice_low = s.AvgPrice
		last_lastPrice_high = s.AvgPrice
		if s.Online {
			// 首單
			var id string
			count := 0
			fmt.Println("首單")
			for {
				id = s.LaunchOrder(wrappers[0], markets[0], s.AvgPrice, &buy_again, "buy")
				count = count + 1
				if count == 20 || id != "" {
					break
				}
			}
			if id != "" {
				fmt.Println("有")
				s.OpenPrice = s.AvgPrice
				s.OrderIdList = append(s.OrderIdList, id)
			} else {
				fmt.Println("無")
				s.UpdateBotInfo("Shut Down Bot!", wrappers[0].Name(), markets[0].Balance)
				isShutDown = true
				s.ShutDown <- true
				return
			}
		}
		s.UpdateBotInfo("First Bot Info!", wrappers[0].Name(), markets[0].Balance)
	}

	if !hasUpdateFunc {
		_err := errors.New("OnUpdate func cannot be empty")
		if hasErrorFunc {
			s.Model.OnError(_err)
		} else {
			panic(_err)
		}
	}
	for err == nil {
		err = s.Model.OnUpdate(wrappers, markets)
		if err != nil && hasErrorFunc {
			s.Model.OnError(err)
		}

		market := markets[0]
		s.Market = market
		lastPrice, _ := market.Summary.Last.Float64()

		//fmt.Println(s.UserId + ": " + s.Model.Name + " " + market.Name + " - AvgPrice: " + fmt.Sprint(s.AvgPrice) + " | lastPrice: " + fmt.Sprint(lastPrice) + " | last_lastPrice_low: " + fmt.Sprint(last_lastPrice_low) + " | last_lastPrice_high: " + fmt.Sprint(last_lastPrice_high))

		if buy_again || sell_again || last_lastPrice_low <= s.AvgPrice*(1.0-s.ActivePercent) || last_lastPrice_high >= s.AvgPrice*(1.0+s.ActivePercent) {

			var id string
			if buy_again || last_lastPrice_low <= s.AvgPrice*(1.0-s.ActivePercent) {
				id = func(last_lastPrice float64, lastPrice float64, AvgPrice float64, GoUpPercent float64) string {
					DropPrice := AvgPrice - last_lastPrice
					GoUpPrice := lastPrice - last_lastPrice
					//fmt.Println(s.UserId + ": " + s.Model.Name + " DropPrice: " + fmt.Sprint(DropPrice) + " | GoUpPrice: " + fmt.Sprint(GoUpPrice))
					if buy_again || (DropPrice > 0 && GoUpPrice > 0 && GoUpPrice >= DropPrice*GoUpPercent) {
						// BuyMarket
						//fmt.Println(s.UserId + ": " + s.Model.Name + " DropPrice*GoUpPercent: " + fmt.Sprint(DropPrice*GoUpPercent))
						return s.LaunchOrder(wrappers[0], market, lastPrice, &buy_again, "buy")
					}
					return ""
				}(last_lastPrice_low, lastPrice, s.AvgPrice, s.ReversePercent)
				if id != "" {
					if len(s.OrderIdList) == 0 {
						s.OpenPrice = s.AvgPrice
					}
					s.OrderIdList = append(s.OrderIdList, id)
					s.AvgPrice = s.UpdateAvgPrice(wrappers[0], market)
					if s.AvgPrice == 0.0 {
						isShutDown = true
					}
					last_lastPrice_low = s.AvgPrice
					s.UpdateBotInfo("Update Bot Info!", wrappers[0].Name(), market.Balance)
				}
			} else if sell_again || last_lastPrice_high >= s.AvgPrice*(1.0+s.ActivePercent) {
				id = func(last_lastPrice float64, lastPrice float64, AvgPrice float64, GoDnPercent float64) string {
					RisePrice := last_lastPrice - AvgPrice
					GoDnPrice := last_lastPrice - lastPrice
					//fmt.Println(s.UserId + ": " + s.Model.Name + " RisePrice: " + fmt.Sprint(RisePrice) + " | GoDnPrice: " + fmt.Sprint(GoDnPrice))
					if sell_again || (RisePrice > 0 && GoDnPrice > 0 && GoDnPrice >= RisePrice*GoDnPercent) {
						// SellMarket
						//fmt.Println(s.UserId + ": " + s.Model.Name + " RisePrice*GoDnPercent: " + fmt.Sprint(RisePrice*GoDnPercent))
						return s.LaunchOrder(wrappers[0], market, lastPrice, &sell_again, "sell")
					}
					return ""
				}(last_lastPrice_high, lastPrice, s.AvgPrice, s.ReversePercent)
				if id != "" {
					if len(s.OrderIdList) == 0 {
						s.OpenPrice = s.AvgPrice
					}
					s.OrderIdList = append(s.OrderIdList, id)
					profit, _ := s.CalculateProfit(s.AvgPrice, lastPrice, id, wrappers[0])
					s.AvgPrice = s.UpdateAvgPrice(wrappers[0], market)
					if s.AvgPrice == 0.0 {
						isShutDown = true
					}
					last_lastPrice_high = s.AvgPrice
					s.UpdateBotInfo("Update Bot Info!", wrappers[0].Name(), market.Balance)
					if s.CycleType == "single" {
						fmt.Println("Profit: " + fmt.Sprint(profit))
						isShutDown = true
						s.ShutDown <- true
						break
					} else {
						fmt.Println("Profit: " + fmt.Sprint(profit))
						s.OrderIdList = nil
					}
				}
			}

		}

		if lastPrice < last_lastPrice_low {
			last_lastPrice_low = lastPrice
		} else if lastPrice > last_lastPrice_high {
			last_lastPrice_high = lastPrice
		}

		//fmt.Println(s.UserId + ": " + s.Model.Name + " | isShutDown: " + fmt.Sprint(isShutDown))

		if isShutDown || <-s.ClosePosition {
			s.ShutDown <- true
		} else {
			s.ClosePosition <- false
			s.ShutDown <- false
		}
		if !ok {
			fmt.Println("Shut Down")
			break
		}
		//fmt.Println("==================================================")
		time.Sleep(s.Interval)
	}
	if hasTearDownFunc {
		err = s.Model.TearDown(wrappers, markets)
		if err != nil && hasErrorFunc {
			s.Model.OnError(err)
		}
	}
}
