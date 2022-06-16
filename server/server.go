package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	helpers "github.com/BurkeyLai/Trading-Bot/server/bot_helpers"
	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/BurkeyLai/Trading-Bot/server/exchanges"
	"github.com/BurkeyLai/Trading-Bot/server/proto"
	"github.com/BurkeyLai/Trading-Bot/server/strategies"
	"github.com/shopspring/decimal"
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

type Server struct {
	Connections map[string]*Connection
	SpotBots    map[string]map[string]map[string]strategies.SpotBotStrategy // map[exchange name]map[market name]strategies.SpotBotStrategy
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
		wrappers[user.Exchcfg.Exchs[i].Exchname] = helpers.InitExchange(exchangeConfig, false, make(map[string]decimal.Decimal), exchangeConfig.DepositAddresses)
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

func (s *Server) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderRespond, error) {
	userId := req.Msg.User.Id
	conn := s.Connections[userId]
	wrapper := conn.exchanges[req.Exchange]

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

	var lotSizeMinQty float64
	var lotSizeMaxQty float64
	var minNotional float64
	markets, err := wrapper.GetMarkets(req.Mode)
	if err != nil {
		fmt.Println(err)
		return &proto.CreateOrderRespond{}, err
	}

	resp := &proto.CreateOrderRespond{
		Timestamp: time.Now().String(),
		//Orderid:   "", // QcpN1VqXk5eUqclJp8phBd, dHWTBEkLGag7VgrUKXvwB6, 6jFjYMLjkZ3ZeqyGiyXe2V, THISO1GxXsemRMdb1eNXMB, iRQnkDYIVu3qohErr2pNxd, t0NZz29Xy55fEW3tDj2PJH, A8r41mK6gFpJ5l0ysvmlcq
		Content:   "",
		Botactive: false,
	}

	quantity, _ := strconv.ParseFloat(req.Quantity, 64)

	for _, market := range markets {
		var symbolSummary *environment.MarketSummary
		var amount float64
		if market.Name == m1.Name {
			lotSizeMinQty, _ = strconv.ParseFloat(market.LotSizeMinQty, 64)
			lotSizeMaxQty, _ = strconv.ParseFloat(market.LotSizeMaxQty, 64)
			minNotional, _ = strconv.ParseFloat(market.MinNotional, 64)
			symbolSummary, err = wrapper.GetMarketSummary(m1)
			if err != nil {
				return &proto.CreateOrderRespond{}, err
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
				resp.Content = "Symbol 1 Amount Size Too Large"
				return resp, nil
			} else if amount < lotSizeMinQty {
				resp.Content = "Symbol 1 Amount Size Too Small"
				return resp, nil
			} else if quantity < minNotional {
				resp.Content = "Symbol 1 Quantity Too Small"
				return resp, nil
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

	if req.Mode == "spot" {
		go func(wrapper exchanges.ExchangeWrapper, m1 *environment.Market) {
			m1.LotSizeMaxQty = fmt.Sprint(lotSizeMaxQty)
			m1.LotSizeMinQty = fmt.Sprint(lotSizeMinQty)
			m1.MinNotional = fmt.Sprint(minNotional)
			wrapperArray := []exchanges.ExchangeWrapper{wrapper}
			marketArray := []*environment.Market{m1}
			exchName := strings.ToLower(req.Exchange)
			symbolName := m1.Name
			bot := SpotBot
			bot.UserName = conn.user.Name
			bot.UserId = conn.user.Id
			bot.Model.Name = symbolName
			bot.DropPercent, _ = strconv.ParseFloat(req.Droppercent, 64)
			//bot.DropPercent = 0.0001
			bot.GoUpPercent, _ = strconv.ParseFloat(req.Gouppercent, 64)
			bot.Qty = quantity
			bot.Stream = conn.stream

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
			s.SpotBots[userId][exchName][symbolName].Apply(wrapperArray, marketArray)
		}(wrapper, m1)
	} else {

	}

	resp.Botactive = true
	return resp, nil
}

func (s *Server) AccountBalance(ctx context.Context, req *proto.AccountBalanceRequest) (*proto.AccountBalanceRespond, error) {
	fmt.Println(req.Msg.Content)
	conn := s.Connections[req.Msg.User.Id]
	wrapper := conn.exchanges[req.Exchange]
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

func main() {

	server := &Server{
		Connections: make(map[string]*Connection),
		SpotBots:    make(map[string]map[string]map[string]strategies.SpotBotStrategy),
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
