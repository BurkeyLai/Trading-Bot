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
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/BurkeyLai/Trading-Bot/server/exchanges"
	"github.com/BurkeyLai/Trading-Bot/server/proto"
)

// SpotBotStrategy is an interval based strategy.
type SpotBotStrategy struct {
	UserName    string
	UserId      string
	Model       StrategyModel
	AvgPrice    float64
	Qty         float64
	DropPercent float64
	GoUpPercent float64
	OrderIdList []string
	Interval    time.Duration
	Stream      proto.TradingBot_CreateStreamServer
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
				TotalUsd += usd
				TotalQty += qty
				break
			}
		}
	}
	AvgPrice := TotalUsd / TotalQty
	return AvgPrice
}

// Apply executes Cyclically the On Update, basing on provided interval.
func (s SpotBotStrategy) Apply(wrappers []exchanges.ExchangeWrapper, markets []*environment.Market) {
	var err error
	var last_lastPrice float64

	hasSetupFunc := s.Model.Setup != nil
	hasTearDownFunc := s.Model.TearDown != nil
	hasUpdateFunc := s.Model.OnUpdate != nil
	hasErrorFunc := s.Model.OnError != nil

	if hasSetupFunc {
		err = s.Model.Setup(wrappers, markets)
		if err != nil && hasErrorFunc {
			s.Model.OnError(err)
		}

		s.AvgPrice, _ = markets[0].Summary.Last.Float64()
		last_lastPrice = s.AvgPrice
		msg := &proto.Message{
			User: &proto.User{
				Id:   s.UserId,
				Name: s.UserName,
			},
			Botinfo: &proto.BotInfo{
				Exch:        wrappers[0].Name(),
				Mode:        "spot",
				Modelname:   s.Model.Name,
				Avgprice:    fmt.Sprint(s.AvgPrice),
				Orderidlist: s.OrderIdList,
			},
			Content:   "First Bot Info!",
			Timestamp: time.Now().String(),
		}
		s.Stream.Send(msg)
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
		lastPrice, _ := market.Summary.Last.Float64()
		fmt.Println(s.Model.Name + " " + market.Name + " - AvgPrice: " + fmt.Sprint(s.AvgPrice) + " | lastPrice: " + fmt.Sprint(lastPrice) + " | last_lastPrice: " + fmt.Sprint(last_lastPrice))
		if last_lastPrice <= s.AvgPrice*(1.0-s.DropPercent) {
			id := func(last_lastPrice float64, lastPrice float64, AvgPrice float64, GoUpPercent float64) string {
				DropPrice := AvgPrice - last_lastPrice
				GoUpPrice := lastPrice - last_lastPrice
				fmt.Println(s.Model.Name + " DropPrice: " + fmt.Sprint(DropPrice) + " | GoUpPrice: " + fmt.Sprint(GoUpPrice))
				if DropPrice > 0 && GoUpPrice > 0 && GoUpPrice >= DropPrice*GoUpPercent {
					// BuyMarket
					fmt.Println(s.Model.Name + " DropPrice*GoUpPercent: " + fmt.Sprint(DropPrice*GoUpPercent))

					amount := s.Qty / lastPrice
					lotSizeMinQty, _ := strconv.ParseFloat(market.LotSizeMinQty, 64)
					pow := 0.0
					for {
						pow++
						lotSizeMinQty *= 10
						if lotSizeMinQty == 1.0 {
							break
						}
					}
					amount = math.Ceil(amount*math.Pow(10, pow)) / math.Pow(10, pow)

					//fmt.Println(pow)
					fmt.Println("quantity: " + fmt.Sprint(s.Qty))
					fmt.Println("lastPrice: " + fmt.Sprint(lastPrice))
					fmt.Println("amount: " + fmt.Sprint(amount))

					clientOrderId, err := wrappers[0].BuyMarket(market, amount)
					if err != nil {
						fmt.Println(err)
						return ""
					}
					//clientOrderId, _ := wrapper.SellMarket(m, amount)
					//resp.Orderid = clientOrderId

					fmt.Println("clientOrderId: === " + clientOrderId + " ===")
					return clientOrderId
				}
				return ""
			}(last_lastPrice, lastPrice, s.AvgPrice, s.GoUpPercent)

			if id != "" {
				s.OrderIdList = append(s.OrderIdList, id)
				s.AvgPrice = s.UpdateAvgPrice(wrappers[0], market)
				last_lastPrice = s.AvgPrice
				msg := &proto.Message{
					User: &proto.User{
						Id:   s.UserId,
						Name: s.UserName,
					},
					Botinfo: &proto.BotInfo{
						Exch:        wrappers[0].Name(),
						Mode:        "spot",
						Modelname:   s.Model.Name,
						Avgprice:    fmt.Sprint(s.AvgPrice),
						Orderidlist: s.OrderIdList,
					},
					Content:   "Update Bot Info!",
					Timestamp: time.Now().String(),
				}
				s.Stream.Send(msg)
			}
		}

		if lastPrice < last_lastPrice {
			last_lastPrice = lastPrice
		}

		fmt.Println("==================================================")
		time.Sleep(s.Interval)
	}
	if hasTearDownFunc {
		err = s.Model.TearDown(wrappers, markets)
		if err != nil && hasErrorFunc {
			s.Model.OnError(err)
		}
	}
}
