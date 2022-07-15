package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	helpers "github.com/BurkeyLai/Trading-Bot/server/bot_helpers"
	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/BurkeyLai/Trading-Bot/server/exchanges"
	"github.com/BurkeyLai/Trading-Bot/server/proto"
	"github.com/BurkeyLai/Trading-Bot/server/strategies"
	"github.com/BurkeyLai/Trading-Bot/server/utils"
	"github.com/shopspring/decimal"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2
var mu sync.Mutex

const (
	WAIT = "wait"
)

func init() {

	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)

}

type Connection struct {
	stream    proto.TradingBot_CreateStreamServer
	user      *proto.User
	active    bool
	exchanges map[string]exchanges.ExchangeWrapper
	error     chan error
}

type SpotManagerFunc func(map[string]map[string]map[string]strategies.SpotBotStrategy) error

type Server struct {
	Connections map[string]*Connection
	SpotBots    map[string]map[string]map[string]strategies.SpotBotStrategy // map[user id]map[exchange name]map[market name]strategies.SpotBotStrategy
	Firestore   *firestore.Client
	SpotManager SpotManagerFunc
}

type BotInfo struct {
	average_price string
	bot_active    bool
	cycle_type    string
	drop_percent  string
	exchange      string
	go_up_percent string
	leverage      string
	max_drawdown  string
	mode          string
	order_id_list []string
	quantity      string
	symbol        string
	m_type        string
	withdraw_spot string
}

var SpotBot = strategies.SpotBotStrategy{
	Model: strategies.StrategyModel{
		//Name: "SpotBot",
		Setup: func(wrappers []exchanges.ExchangeWrapper, markets []*environment.Market) error {
			fmt.Println("SpotBot starting")
			for _, market := range markets {
				summary, err := wrappers[0].GetMarketSummary(market)
				if err != nil {
					return err
				}
				market.Summary = *summary

				requestSymbol := market.BaseCurrency
				if wrappers[0].Name() == "binance" {
					requestSymbol = strings.ToUpper(requestSymbol)
				}
				b, err := wrappers[0].GetBalance("spot", requestSymbol)
				if err != nil {
					return err
				}
				market.Balance = b.String()
				//balance, _ := strconv.ParseFloat(b.String(), 64)
			}
			return nil
		},
		OnUpdate: func(wrappers []exchanges.ExchangeWrapper, markets []*environment.Market) error {
			for _, market := range markets {
				summary, err := wrappers[0].GetMarketSummary(market)
				if err != nil {
					return err
				}
				market.Summary = *summary

				requestSymbol := market.BaseCurrency
				if wrappers[0].Name() == "binance" {
					requestSymbol = strings.ToUpper(requestSymbol)
				}
				b, err := wrappers[0].GetBalance("spot", requestSymbol)
				if err != nil {
					return err
				}
				market.Balance = b.String()

				//orderbook, err := wrappers[0].GetOrderBook(market)
				//if err != nil {
				//	return err
				//}
				//market.Orders = *orderbook

				//orders, err := wrappers[0].AskOrderList("spot", market)
				//if err != nil {
				//	return err
				//}
				//market.Orders = orders
			}
			//logrus.Info(markets)
			//logrus.Info(wrappers)
			return nil
		},
		OnError: func(err error) {
			fmt.Println(err)
		},
		TearDown: func(wrappers []exchanges.ExchangeWrapper, markets []*environment.Market) error {
			fmt.Println("SpotBot exited")
			return nil
		},
	},
	Interval: time.Second * 5,
}

func (s *Server) ExecuteStartCommand(user *proto.User) map[string]exchanges.ExchangeWrapper {
	fmt.Print("Getting exchange info ... ")

	exchnum := len(user.Exchcfg.Exchs)
	wrappers := make(map[string]exchanges.ExchangeWrapper)
	for i := 0; i < exchnum; i++ {
		fmt.Println()
		fmt.Println(user.Exchcfg.Exchs[i].Exchname)
		fmt.Println(user.Exchcfg.Exchs[i].Spotpublickey)
		fmt.Println(user.Exchcfg.Exchs[i].Spotsecretkey)
		exchangeConfig := environment.ExchangeConfig{
			ExchangeName:     user.Exchcfg.Exchs[i].Exchname,
			SpotPublicKey:    user.Exchcfg.Exchs[i].Spotpublickey,
			SpotSecretKey:    user.Exchcfg.Exchs[i].Spotsecretkey,
			FuturePublicKey:  user.Exchcfg.Exchs[i].Futurepublickey,
			FutureSecretKey:  user.Exchcfg.Exchs[i].Futuresecretkey,
			DepositAddresses: make(map[string]string),
		}
		exchangeConfig.DepositAddresses["BTC"] = user.Exchcfg.Exchs[i].Depoaddr.Btcaddr
		exchangeConfig.DepositAddresses["ETH"] = user.Exchcfg.Exchs[i].Depoaddr.Ethaddr
		if exchangeConfig.SpotPublicKey != "" && exchangeConfig.SpotSecretKey != "" {
			wrappers[user.Exchcfg.Exchs[i].Exchname] = helpers.InitExchange(exchangeConfig, false, make(map[string]decimal.Decimal), exchangeConfig.DepositAddresses)
		}
	}
	fmt.Println("DONE")

	return wrappers
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.TradingBot_CreateStreamServer) error {
	conn := &Connection{
		stream:    stream,
		user:      pconn.User,
		active:    true,
		exchanges: make(map[string]exchanges.ExchangeWrapper),
		error:     make(chan error),
	}
	conn.exchanges = s.ExecuteStartCommand(conn.user)
	s.Connections[conn.user.Id] = conn

	msg := &proto.Message{
		User: &proto.User{
			Id:   conn.user.Id,
			Name: conn.user.Name,
		},
		Content:   "Create stream finished!",
		Timestamp: time.Now().String(),
	}
	stream.Send(msg)

	return <-conn.error
}

func (s *Server) MarketInfo(ctx context.Context, req *proto.MarketInfoRequest) (*proto.MarketInfoRespond, error) {
	fmt.Println(req.Msg.Content)
	//fmt.Println(req.Msg.User.Id)
	//fmt.Println(req.Exchange)
	conn := s.Connections[req.Msg.User.Id]
	wrapper := conn.exchanges[req.Exchange]
	if wrapper == nil {
		return &proto.MarketInfoRespond{}, nil
	}
	markets, err := wrapper.GetMarkets(req.Mode)
	if err != nil {
		fmt.Println(err)
		return &proto.MarketInfoRespond{}, err
	}

	var requestType string
	if req.Type == "usdt" {
		if req.Exchange == "Huobi" {
			requestType = "usdt"
		} else if req.Exchange == "Binance" {
			requestType = "USDT"
		}
	}

	symbols := []string{}
	for _, market := range markets {
		if market.MarketCurrency == requestType {

			if req.Mode == "spot" {
				// If the bot serve that symbol is ignited, don't show that symbol
				userExch := s.SpotBots[req.Msg.User.Id]
				exchMarkets := userExch[strings.ToLower(req.Exchange)]
				_, botExist := exchMarkets[market.Name]
				if !botExist {
					symbols = append(symbols, market.BaseCurrency+"/"+market.MarketCurrency)
				}
			} else {

			}
			//fmt.Printf("{" + market.Name + ": [" + market.BaseCurrency + ", " + market.MarketCurrency + "]} ")

		}
		//fmt.Println(market.MarketCurrency)
	}
	sort.Strings(symbols)
	resp := &proto.MarketInfoRespond{
		Timestamp: time.Now().String(),
		Symbols:   symbols,
	}

	info := &proto.Message{
		User: &proto.User{
			Id:   conn.user.Id,
			Name: conn.user.Name,
		},
		Content:   "Info received!",
		Timestamp: time.Now().String(),
	}
	conn.stream.Send(info)
	return resp, nil
}

func (s *Server) LaunchBot(
	wrapper exchanges.ExchangeWrapper,
	mode,
	qty,
	exchange,
	dropPercent,
	goUpPercent,
	cycleType,
	userId,
	userName string,
	orderList []string,
	online bool,
	m1 *environment.Market,
	stream *proto.TradingBot_CreateStreamServer,
	doc *firestore.DocumentRef) (string, error) {

	var lotSizeMinQty float64
	var lotSizeMaxQty float64
	var minNotional float64
	markets, err := wrapper.GetMarkets(mode)
	if err != nil {
		fmt.Println(err)
		return "GetMarkets Error!", err
	}
	quantity, _ := strconv.ParseFloat(qty, 64)

	for _, market := range markets {
		var symbolSummary *environment.MarketSummary
		var amount float64
		if market.Name == m1.Name {
			lotSizeMinQty, _ = strconv.ParseFloat(market.LotSizeMinQty, 64)
			lotSizeMaxQty, _ = strconv.ParseFloat(market.LotSizeMaxQty, 64)
			minNotional, _ = strconv.ParseFloat(market.MinNotional, 64)
			symbolSummary, err = wrapper.GetMarketSummary(m1)
			if err != nil {
				return "GetMarketSummary Error!", err
			}
			lastPrice, _ := symbolSummary.Last.Float64()
			amount = quantity / lastPrice

			//fmt.Println("quantity: " + fmt.Sprint(quantity))
			//fmt.Println("lastPrice: " + fmt.Sprint(lastPrice))
			//fmt.Println("amount: " + fmt.Sprint(amount))
			//fmt.Println("lotSizeMaxQty: " + fmt.Sprint(lotSizeMaxQty))
			//fmt.Println("lotSizeMinQty: " + fmt.Sprint(lotSizeMinQty))
			//fmt.Println("minNotional: " + fmt.Sprint(minNotional))

			if amount > lotSizeMaxQty {
				content := "Symbol 1 Amount Size Too Large"
				return content, errors.New(content)
			} else if amount < lotSizeMinQty {
				content := "Symbol 1 Amount Size Too Small"
				return content, errors.New(content)
			} else if quantity < minNotional {
				content := "Symbol 1 Quantity Too Small"
				return content, errors.New(content)
			}
		}

		fmt.Println()
		//fmt.Println(symbolSummary.Last)
		//fmt.Println(symbolSummary.Ask)
		//fmt.Println(symbolSummary.Bid)
		//fmt.Println(symbolSummary.High)
		//fmt.Println(symbolSummary.Low)
		//fmt.Println(symbolSummary.Volume)
		//fmt.Println(m.Name)
		//fmt.Println(m.BaseCurrency)
		//fmt.Println(m.MarketCurrency)
		//fmt.Println(strings.ToLower(req.Exchange))
		//fmt.Println(lotSizeMinQty)
		//fmt.Println(lotSizeMaxQty)
	}

	if mode == "spot" {
		mu.Lock()
		//go func(wrapper exchanges.ExchangeWrapper, m1 *environment.Market) {
		func(wrapper exchanges.ExchangeWrapper, m1 *environment.Market) {
			m1.LotSizeMaxQty = fmt.Sprint(lotSizeMaxQty)
			m1.LotSizeMinQty = fmt.Sprint(lotSizeMinQty)
			m1.MinNotional = fmt.Sprint(minNotional)
			wrapperArray := []exchanges.ExchangeWrapper{wrapper}
			marketArray := []*environment.Market{m1}
			exchName := strings.ToLower(exchange)
			symbolName := m1.Name
			bot := SpotBot
			bot.UserName = userName
			bot.UserId = userId
			bot.CycleType = cycleType
			bot.Model.Name = symbolName
			bot.ActivePercent, _ = strconv.ParseFloat(dropPercent, 64)
			//bot.DropPercent = 0.0001
			bot.ReversePercent, _ = strconv.ParseFloat(goUpPercent, 64)
			bot.Qty = quantity
			if online {
				bot.Stream = *stream
			} else {
				bot.OrderIdList = orderList
				bot.Doc = doc
			}
			bot.Market = m1
			bot.Online = online
			bot.ClosePosition = make(chan bool, 1)
			bot.ShutDown = make(chan bool, 1)

			var exchBots map[string]strategies.SpotBotStrategy
			_, exchBotsExist := s.SpotBots[userId][exchName]
			if exchBotsExist {
				exchBots = s.SpotBots[userId][exchName]
			} else {
				exchBots = make(map[string]strategies.SpotBotStrategy)
			}
			exchBots[symbolName] = bot

			var userExchs map[string]map[string]strategies.SpotBotStrategy
			_, userExchsExist := s.SpotBots[userId]
			if userExchsExist {
				userExchs = s.SpotBots[userId]
			} else {
				userExchs = make(map[string]map[string]strategies.SpotBotStrategy)
			}
			userExchs[exchName] = exchBots

			s.SpotBots[userId] = userExchs

			go s.SpotBots[userId][exchName][symbolName].Apply(wrapperArray, marketArray)
		}(wrapper, m1)
		mu.Unlock()
	} else {

	}

	return "LaunchBot Success!", nil
}

func (s *Server) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderRespond, error) {
	userId := req.Msg.User.Id
	userName := req.Msg.User.Name
	conn := s.Connections[userId]
	wrapper := conn.exchanges[req.Exchange]
	if wrapper == nil {
		return &proto.CreateOrderRespond{}, nil
	}

	requestSymbol1 := req.Symbol
	//requestSymbol2 := req.Symbol2
	if req.Exchange == "Binance" {
		requestSymbol1 = strings.ToUpper(requestSymbol1)
		//requestSymbol2 = strings.ToUpper(requestSymbol2)
	}
	if requestSymbol1 == "" {
		return &proto.CreateOrderRespond{}, nil
	}
	strs1 := strings.Split(requestSymbol1, "/")
	m1 := &environment.Market{
		Name:           strs1[0] + strs1[1],
		BaseCurrency:   strs1[0],
		MarketCurrency: strs1[1],
		ExchangeNames:  make(map[string]string),
	}
	m1.ExchangeNames[strings.ToLower(req.Exchange)] = m1.Name

	//var m2 *environment.Market
	//if requestSymbol2 == "" {
	//	m2 = nil
	//} else {
	//	strs2 := strings.Split(requestSymbol2, "/")
	//	m2 = &environment.Market{
	//		Name:           strs2[0] + strs2[1],
	//		BaseCurrency:   strs2[0],
	//		MarketCurrency: strs2[1],
	//		ExchangeNames:  make(map[string]string),
	//	}
	//	m2.ExchangeNames[strings.ToLower(req.Exchange)] = m2.Name
	//}
	resp := &proto.CreateOrderRespond{
		Timestamp: time.Now().String(),
		//Orderid:   "", // QcpN1VqXk5eUqclJp8phBd, dHWTBEkLGag7VgrUKXvwB6, 6jFjYMLjkZ3ZeqyGiyXe2V, THISO1GxXsemRMdb1eNXMB, iRQnkDYIVu3qohErr2pNxd, t0NZz29Xy55fEW3tDj2PJH, A8r41mK6gFpJ5l0ysvmlcq
		Content:   "",
		Botactive: false,
	}
	content, err := s.LaunchBot(
		wrapper,
		req.Mode,
		req.Quantity,
		req.Exchange,
		req.Droppercent,
		req.Gouppercent,
		req.Cycletype,
		userId,
		userName,
		[]string{},
		true,
		m1,
		&conn.stream,
		nil)
	resp.Content = content
	if err != nil {
		return resp, err
	}

	resp.Botactive = true
	return resp, nil
}

func (s *Server) AccountBalance(ctx context.Context, req *proto.AccountBalanceRequest) (*proto.AccountBalanceRespond, error) {
	fmt.Println(req.Msg.Content)
	conn := s.Connections[req.Msg.User.Id]
	wrapper := conn.exchanges[req.Exchange]
	if wrapper == nil || req.Mode == "" || req.Symbol == "" {
		return &proto.AccountBalanceRespond{}, nil
	}
	requestSymbol := req.Symbol
	if req.Exchange == "Binance" {
		requestSymbol = strings.ToUpper(requestSymbol)
	}

	balance, err := wrapper.GetBalance(req.Mode, requestSymbol)
	if err != nil {
		return &proto.AccountBalanceRespond{}, err
	}

	resp := &proto.AccountBalanceRespond{
		Timestamp: time.Now().String(),
		Balance:   balance.String(),
	}

	//fmt.Println(balance)

	return resp, nil
}

func (s *Server) OrderInfo(ctx context.Context, req *proto.OrderInfoRequest) (*proto.OrderInfoRespond, error) {
	fmt.Println(req.Msg.Content)
	conn := s.Connections[req.Msg.User.Id]
	wrapper := conn.exchanges[req.Exchange]
	if wrapper == nil || req.Mode == "" || req.Symbol == "" || req.Orderid == "" {
		return &proto.OrderInfoRespond{}, nil
	}
	requestSymbol := req.Symbol
	if req.Exchange == "Binance" {
		requestSymbol = strings.ToUpper(requestSymbol)
	}

	//fmt.Println(req)
	order, err := wrapper.QueryOrder(req.Mode, req.Orderid, requestSymbol)
	if err != nil {
		return &proto.OrderInfoRespond{}, err
	}

	if order == nil {
		return &proto.OrderInfoRespond{}, nil
	}
	qty, _ := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)
	amount, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	price := fmt.Sprint(qty / amount)

	resp := &proto.OrderInfoRespond{
		Timestamp: fmt.Sprint(time.Unix(order.Time/1000, 0)),
		Quantity:  order.CummulativeQuoteQuantity,
		Amount:    order.ExecutedQuantity,
		Price:     price,
		Type:      string(order.Type),
		Side:      string(order.Side),
	}

	//fmt.Println(balance)

	return resp, nil
}

func (s *Server) ClosePosition(ctx context.Context, req *proto.ClosePositionRequest) (*proto.ClosePositionRespond, error) {
	userid := req.Msg.User.Id
	exchange := req.Exchange
	conn := s.Connections[userid]
	wrapper := conn.exchanges[exchange]
	if wrapper == nil || req.Mode == "" || req.Symbol == "" {
		return &proto.ClosePositionRespond{}, nil
	}
	requestSymbol := req.Symbol
	if exchange == "Binance" {
		requestSymbol = strings.ToUpper(requestSymbol)
	}

	var clientOrderId string
	if req.Mode == "spot" {
		//bot := s.SpotBots[userid][strings.ToLower(exchange)][requestSymbol]
		userExch := s.SpotBots[userid]
		exchMarkets := userExch[strings.ToLower(exchange)]
		bot, botExist := exchMarkets[requestSymbol]

		if botExist {
			market := bot.Market
			lotSizeMinQty, _ := strconv.ParseFloat(market.LotSizeMinQty, 64)
			minNotional, _ := strconv.ParseFloat(market.MinNotional, 64)
			TotalQty := 0.0
			orders, err := wrapper.AskOrderList("spot", market)
			if err != nil {
				fmt.Println(err)
			}
			market.Orders = orders
			for _, orderId := range bot.OrderIdList {
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
			price, _ := market.Summary.Last.Float64()
			amount, err := utils.CorrectPrecision(TotalQty, price, lotSizeMinQty, minNotional)
			if err != nil {
				return &proto.ClosePositionRespond{}, nil
			}
			for {
				clientOrderId, err = wrapper.SellMarket(market, amount)
				if err != nil {
					fmt.Println(err)
					return &proto.ClosePositionRespond{}, err
				}
				if clientOrderId != "" {
					break
				}
			}
			profit := (price - bot.AvgPrice) * TotalQty
			fmt.Println("Profit: " + fmt.Sprint(profit))

			bot.ClosePosition <- true
		} else {
			return &proto.ClosePositionRespond{}, nil
		}

	} else {

	}

	resp := &proto.ClosePositionRespond{
		Timestamp: time.Now().String(),
		Content:   clientOrderId,
		Botactive: false,
	}

	return resp, nil
}

func main() {
	ctx := context.Background()
	sa := option.WithCredentialsFile("../trading-bot-d40d7-firebase-adminsdk-nnagv-d8b080c313.json") // Firebase -> 專案設定 -> 服務帳戶 -> 產生新的私密金鑰
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	server := &Server{
		Connections: make(map[string]*Connection),
		SpotBots:    make(map[string]map[string]map[string]strategies.SpotBotStrategy),
		Firestore:   client,
		SpotManager: func(bots map[string]map[string]map[string]strategies.SpotBotStrategy) error {
			var err error

			copy_bots := bots
			for err == nil {
				wait1 := sync.WaitGroup{}
				//done1 := make(chan int)
				wait1.Add(1)

				//fmt.Println("-----------------------")
				go func() {
					defer wait1.Done()
					var id, exch, symbol string
					for i, exchs := range copy_bots {
						//fmt.Println("user id: " + i)
						for e, symbols := range exchs {
							//fmt.Println("exchange: " + e)
							for s, bot := range symbols {
								sd := <-bot.ShutDown
								//fmt.Println("symbol: " + s)
								//fmt.Print("ShutDown: ")
								//fmt.Println(sd)

								if sd {
									//delete(symbols, symbol)
									id = i
									exch = e
									symbol = s
									close(bot.ShutDown)
									fmt.Println("++++++++++++++++++++++")
									fmt.Println(bot.Model.Name)
									fmt.Println("++++++++++++++++++++++")
								}
							}
						}
					}
					//delete(bots[id][exch], symbol)

					userExch := bots[id]
					exchMarkets := userExch[exch]
					delete(exchMarkets, symbol)

				}()
				go func() {
					wait1.Wait()
					//	close(done1)
				}()
				//<-done1

				//fmt.Println("-----------------------")
				time.Sleep(time.Second * 5)
			}

			return nil

		},
	}

	go server.SpotManager(server.SpotBots)

	iter := client.Collection("User").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		//fmt.Println(doc.DataAt("email"))
		//fmt.Println(doc.Ref.ID)
		//fmt.Println(doc.DataAt("name"))
		exchCfg := &proto.ExchangeConfig{
			Exchs: []*proto.Exchange{},
		}
		for i := 0; i < 2; i++ { // 2: number of exchanges -> binance, huobi
			depositAddr := &proto.DepositAddresses{
				Addrnum: "2",
				Btcaddr: "",
				Ethaddr: "",
			}
			exch := &proto.Exchange{
				Depoaddr: depositAddr,
			}
			if i == 0 { // huobi
				//exch.Exchname = "Huobi"
			} else if i == 1 { // binance
				exch.Exchname = "Binance"
				binance_spot_apikey, err := doc.DataAt("binance_spot_apikey")
				if binance_spot_apikey != nil && err == nil {
					switch t := binance_spot_apikey.(type) {
					case string:
						exch.Spotpublickey = t
					}
				}
				binance_spot_secretkey, err := doc.DataAt("binance_spot_secretkey")
				if binance_spot_secretkey != nil && err == nil {
					switch t := binance_spot_secretkey.(type) {
					case string:
						exch.Spotsecretkey = t
					}
				}
				binance_future_apikey, err := doc.DataAt("binance_future_apikey")
				if binance_future_apikey != nil && err == nil {
					switch t := binance_future_apikey.(type) {
					case string:
						exch.Futurepublickey = t
					}
				}
				binance_future_secretkey, err := doc.DataAt("binance_future_secretkey")
				if binance_future_secretkey != nil && err == nil {
					switch t := binance_future_secretkey.(type) {
					case string:
						exch.Futuresecretkey = t
					}
				}
			}

			exchCfg.Exchs = append(exchCfg.Exchs, exch)
		}
		user := &proto.User{
			Id:      doc.Ref.ID,
			Exchcfg: exchCfg,
		}

		//if user.Id != "8GvKxUkTqfMpQMWTaVizOLxw9sj2" {
		//	continue
		//}

		name, err := doc.DataAt("name")
		if name != nil && err == nil {
			switch t := name.(type) {
			case string:
				user.Name = t
			}
		}

		wrappers := server.ExecuteStartCommand(user)
		bots_array, err := doc.DataAt("bots_array")
		if bots_array != nil {
			switch t := bots_array.(type) {
			case []interface{}:
				for _, bot := range t {
					botInfo := &BotInfo{}
					val := reflect.ValueOf(bot)
					if val.Kind() == reflect.Map {
						for _, key := range val.MapKeys() {
							v := val.MapIndex(key)
							switch value := v.Interface().(type) {
							case string:
								switch key.String() {
								case "average_price":
									botInfo.average_price = value
								case "cycle_type":
									botInfo.cycle_type = value
								case "drop_percent":
									botInfo.drop_percent = value
								case "exchange":
									botInfo.exchange = value
								case "go_up_percent":
									botInfo.go_up_percent = value
								case "leverage":
									botInfo.leverage = value
								case "max_drawdown":
									botInfo.max_drawdown = value
								case "mode":
									botInfo.mode = value
								case "quantity":
									botInfo.quantity = value
								case "symbol":
									botInfo.symbol = value
								case "m_type":
									botInfo.m_type = value
								case "withdraw_spot":
									botInfo.withdraw_spot = value
								}
								//fmt.Println(key.String(), value)
							case bool:
								switch key.String() {
								case "bot_active":
									botInfo.bot_active = value
								}
								//fmt.Println(key.String(), value)
							case []interface{}:
								//fmt.Println(key.String())
								switch key.String() {
								case "order_id_list":
									for _, item := range value {
										switch id := item.(type) {
										case string:
											//fmt.Println(id)
											botInfo.order_id_list = append(botInfo.order_id_list, id)
										}
									}
								}
							}
						}
						//fmt.Println(val.MapIndex(reflect.ValueOf("cycle_type")))
						//fmt.Println(val.MapIndex(reflect.ValueOf("average_price")))
						//val.MapIndex(reflect.ValueOf("cycle_type")).Set(reflect.ValueOf(fmt.Sprint("gg")))
						//val.SetMapIndex(reflect.ValueOf("cycle_type"), reflect.ValueOf("gg"))
						//fmt.Println(val.MapIndex(reflect.ValueOf("cycle_type")))
						//client.Collection("User").Doc(user.Id).Set(ctx, map[string]interface{}{
						//	"bots_array": bots_array,
						//}, firestore.MergeAll)
					}
					//fmt.Println(bot)
					//fmt.Println(botInfo)

					var wrapper exchanges.ExchangeWrapper
					if botInfo.exchange == "binance" {
						wrapper = wrappers["Binance"]
					} else if botInfo.exchange == "huobi" {
						wrapper = wrappers["Huobi"]
					}
					var requestType string
					if botInfo.m_type == "usdt" {
						if botInfo.exchange == "huobi" {
							requestType = "usdt"
						} else if botInfo.exchange == "binance" {
							requestType = "USDT"
						}
					}
					m1 := &environment.Market{
						Name:           botInfo.symbol,
						BaseCurrency:   strings.ReplaceAll(botInfo.symbol, requestType, ""),
						MarketCurrency: requestType,
						ExchangeNames:  make(map[string]string),
					}
					m1.ExchangeNames[botInfo.exchange] = m1.Name

					server.LaunchBot(
						wrapper,
						botInfo.mode,
						botInfo.quantity,
						botInfo.exchange,
						botInfo.drop_percent,
						botInfo.go_up_percent,
						botInfo.cycle_type,
						user.Id,
						user.Name,
						botInfo.order_id_list,
						false,
						m1,
						nil,
						client.Collection("User").Doc(user.Id))

				}
			}
		}
	}

	//init_conn := make(*Connection)
	//server.Connections["Normal"] = init_conn

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	//grpcLog.Info("Starting server at port :8080")
	grpcLog.Info("Starting server at port :9090")

	proto.RegisterTradingBotServer(grpcServer, server)
	grpcServer.Serve(listener)
}

/*
var grpcLog glog.LoggerV2
var redisClient *redis.Client
var mu sync.Mutex

//var server *Server

const (
	WAIT = "wait"
)

func init() {

	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)

	redisClient = redis.NewClient(&redis.Options{
		//Addr: "localhost:6379",
		Addr: "redis:6379",
		//Password: "p@ssword",
		DB: 0, // use default DB
	})
	pong, err := redisClient.Ping(context.Background()).Result()
	if err == nil {
		log.Println("redis 回應成功，", pong)
	} else {
		log.Fatal("redis 無法連線，錯誤為", err)
	}

}


type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	//id      string
	user    *proto.User
	channel *proto.Channel
	tunnel  *Tunnel
	active  bool
	error   chan error
}

type Server struct {
	Channel map[string]map[string]*Tunnel
}

func (s *Server) createTunnel(conn *Connection) error {
	var createTunnelError error
	{
		mu.Lock()
		if key, err := GetWaitFirstKey(conn.channel.Name); err == nil && key != "" {
			CreateChat(conn.user.Id, key)

			userMap := s.Channel[conn.channel.Name]
			userTunnel := userMap[key]
			go userTunnel.RunTunnel()
			userTunnel.register <- userTunnel.creator
			userTunnel.register <- conn
			userMap[conn.user.Id] = userTunnel
			//fmt.Printf("%p\n", userMap[conn.id])
		} else {
			AddToWaitList(conn.channel.Name, conn.user.Id)
			userMap := s.Channel[conn.channel.Name]
			userMap[conn.user.Id] = NewTunnel(conn)
			//conn.tunnel = userMap[conn.id]
			//fmt.Printf("%p\n", userMap[conn.id])
		}
		//go tunnel.RunTunnel()
		//s.rooms[tunnel] = true
		mu.Unlock()
	}

	return createTunnelError
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream:  stream,
		channel: pconn.Channel,
		//id:      pconn.User.Id,
		user:   pconn.User,
		active: true,
		error:  make(chan error),
	}

	//msg := &proto.Message{
	//	Id:        conn.id,
	//	Content:   "對方已經加入聊天室",
	//	Timestamp: time.Now().String(),
	//	Channel:   conn.channel,
	//}
	//stream.Send(msg)

	err := s.createTunnel(conn)
	if err != nil {
		fmt.Printf("Create Tunnel Error: %v", err)
	}

	//c := make(chan *Connection, 1) // https://segmentfault.com/a/1190000021600937
	//c <- conn
	//s.Channel[pconn.Channel.Name] = append(s.Channel[pconn.Channel.Name], conn)
	//fmt.Println(pconn.User.Name + " join " + (s.Channel[pconn.Channel.Name][len(s.Channel[pconn.Channel.Name])-1]).channel.Name)

	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait1 := sync.WaitGroup{}
	done1 := make(chan int)

	wait1.Add(1)
	go func(msg *proto.Message) {
		defer wait1.Done()
		userMap, chanIsExist := s.Channel[msg.Channel.Name]
		//fmt.Println(msg.Channel.Name)
		if chanIsExist {

			tunnel, tunnelIsExist := userMap[msg.Id]
			if msg.Content == "LEAVE" {

				chatTo, _ := redisClient.Get(context.TODO(), msg.Id).Result()
				//fmt.Println("Chat to: " + chatTo)
				//fmt.Println(tunnelIsExist)
				if tunnelIsExist {
					for conn, active := range tunnel.clients {
						if active {
							tunnel.unregister <- conn
							delete(userMap, conn.user.Id)
						}
					}
				}
				if chatTo != "" {
					RemoveChat(msg.Id, chatTo)
				} else {
					RemovePoorBoy(msg.Channel.Name)
				}

			} else {
				tunnel.broadcast <- msg
			}

		}

	}(msg)

	go func() {
		wait1.Wait()
		close(done1)
	}()
	<-done1
	return &proto.Close{}, nil
}

func main() {

	server := &Server{make(map[string]map[string]*Tunnel)}
	init_conn := make(map[string]*Tunnel)
	server.Channel["Normal"] = init_conn

	grpcServer := grpc.NewServer()
	//listener, err := net.Listen("tcp", ":8080")
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	//grpcLog.Info("Starting server at port :8080")
	grpcLog.Info("Starting server at port :9090")

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}

func AddToWaitList(channel, id string) error {
	return redisClient.LPush(context.Background(), channel, id).Err()
}

func GetWaitFirstKey(channel string) (string, error) {
	return redisClient.LPop(context.Background(), channel).Result()
}

func CreateChat(id1, id2 string) {
	redisClient.Set(context.Background(), id1, id2, 0)
	redisClient.Set(context.Background(), id2, id1, 0)
}

func RemoveChat(id1, id2 string) {
	redisClient.Del(context.Background(), id1, id2)
}

func RemovePoorBoy(channel string) {
	key, err := GetWaitFirstKey(channel)
	if err != nil {
		fmt.Println(key)
		fmt.Println(err)
	}
}
*/
