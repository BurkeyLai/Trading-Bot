package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/BurkeyLai/Trading-Bot/server/environment"
	"github.com/BurkeyLai/Trading-Bot/server/exchanges"
	"github.com/BurkeyLai/Trading-Bot/server/proto"
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
	exchanges map[string]*exchanges.HuobiWrapper //map[string]exchanges.ExchangeWrapper
	error     chan error
}

type Server struct {
	Connections map[string]*Connection
}

func executeStartCommand() *exchanges.HuobiWrapper {
	fmt.Print("Getting exchange info ... ")
	//wrappers := make([]exchanges.ExchangeWrapper, len(botConfig.ExchangeConfigs))
	//for i, config := range botConfig.ExchangeConfigs {
	//	wrappers[i] = helpers.InitExchange(config, botConfig.SimulationModeOn, config.FakeBalances, config.DepositAddresses)
	//}
	exchangeConfig := &environment.ExchangeConfig{
		PublicKey:        "12e1r12e",
		SecretKey:        "efwefwef",
		DepositAddresses: make(map[string]string),
	}
	exchangeConfig.DepositAddresses["BTC"] = "kfjspar3"

	var exch *exchanges.HuobiWrapper
	exch = exchanges.NewHuobiWrapper(exchangeConfig.PublicKey, exchangeConfig.SecretKey, exchangeConfig.DepositAddresses)

	fmt.Println("DONE")

	return exch
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.TradingBot_CreateStreamServer) error {
	conn := &Connection{
		stream:    stream,
		user:      pconn.User,
		active:    true,
		exchanges: make(map[string]*exchanges.HuobiWrapper),
		error:     make(chan error),
	}
	conn.exchanges["Huobi"] = executeStartCommand()
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

func (s *Server) MarketInfo(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	fmt.Println(msg.Content)
	conn := s.Connections[msg.User.Id]
	wrapper := conn.exchanges["Huobi"]
	markets, err := wrapper.GetMarkets()
	if err != nil {
		return &proto.Close{}, err
	}

	for _, market := range markets {
		fmt.Printf("{" + market.Name + ": [" + market.BaseCurrency + ", " + market.MarketCurrency + "]} ")
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

	return &proto.Close{}, nil
}

func main() {

	server := &Server{make(map[string]*Connection)}
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
