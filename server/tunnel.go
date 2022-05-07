package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/BurkeyLai/gotunnel/server/proto"

	"github.com/google/uuid"
)

var client proto.BroadcastClient

const (
	MAX_TUNNEL_CLIENTS = 2
)

type Tunnel struct {
	ID uuid.UUID //`json:"id"`
	//Name       string    //`json:"name"`
	creator     *Connection
	num_clients int
	clients     map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	broadcast   chan *proto.Message
}

func NewTunnel(conn *Connection) *Tunnel {
	T := Tunnel{
		ID: uuid.New(),
		//Name:       name,
		creator:     conn,
		num_clients: 0,
		clients:     make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan *proto.Message),
		//Private:    private,
	}

	//T.clients[conn] = true

	return &T
}

// RunTunnel runs our tunnel, accepting various requests
func (tunnel *Tunnel) RunTunnel() {
	for {
		select {

		case conn := <-tunnel.register:
			tunnel.registerClientInTunnel(conn)

		case conn := <-tunnel.unregister:
			tunnel.unregisterClientInTunnel(conn)

		case message := <-tunnel.broadcast:
			tunnel.broadcastToClientsInTunnel(message)
		}

	}
}

func (tunnel *Tunnel) registerClientInTunnel(conn *Connection) {

	if tunnel.num_clients >= MAX_TUNNEL_CLIENTS {
		fmt.Println("////////////////////")
		fmt.Println("// Register Error //")
		fmt.Println("////////////////////")
	} else {

		conn.tunnel = tunnel
		tunnel.clients[conn] = true
		tunnel.num_clients++
		//fmt.Printf("%p\n", tunnel)

		if tunnel.num_clients == MAX_TUNNEL_CLIENTS {
			tunnel.notifyClientJoined(conn)
		}
	}
}

func (tunnel *Tunnel) unregisterClientInTunnel(conn *Connection) {
	if tunnel.num_clients == MAX_TUNNEL_CLIENTS {
		tunnel.notifyClientLeft(conn)
	}
	tunnel.num_clients--
	delete(tunnel.clients, conn)
}

func (tunnel *Tunnel) broadcastToClientsInTunnel(message *proto.Message) (*proto.Close, error) {
	var sendmsgerror error
	//if tunnel.num_clients == MAX_TUNNEL_CLIENTS {

	wait1 := sync.WaitGroup{}
	done1 := make(chan int)

	//var sender, receiver *Connection
	//for conn, active := range tunnel.clients {
	//fmt.Println(tunnel.clients)
	for conn := range tunnel.clients {
		fmt.Println(conn.user.Name)
		//if conn.id == message.Id && active {
		//	sender = conn
		//} else {
		//	receiver = conn
		//}
		wait1.Add(1)
		go func(msg *proto.Message, conn *Connection) {
			defer wait1.Done()
			if conn.active {

				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
					sendmsgerror = err
				}
			}
		}(message, conn)
	}
	//fmt.Println(sender.id + " sends message to " + receiver.id)

	go func() {
		wait1.Wait()
		close(done1)
	}()
	<-done1
	//}
	return &proto.Close{}, sendmsgerror
}

func (tunnel *Tunnel) notifyClientJoined(conn *Connection) {
	msg := &proto.Message{
		//Id:        conn.user.Id,
		Id:        uuid.New().String(),
		Name:      conn.user.Name,
		Content:   "對方已經加入聊天室",
		Timestamp: time.Now().String(),
		Channel:   conn.channel,
		Tunnel: &proto.Tunnel{
			Id:    tunnel.ID.String(),
			User1: tunnel.creator.user.Id,
			User2: conn.user.Id,
		},
	}

	//fmt.Println(tunnel.creator.id)
	//fmt.Println(conn.id)
	fmt.Println(msg)
	tunnel.broadcastToClientsInTunnel(msg)
}

func (tunnel *Tunnel) notifyClientLeft(conn *Connection) {
	msg := &proto.Message{
		//Id:        conn.user.Id,
		Id:        uuid.New().String(),
		Name:      conn.user.Name,
		Content:   "對方已經離開聊天室",
		Timestamp: time.Now().String(),
		Channel:   conn.channel,
	}

	fmt.Println(msg)
	tunnel.broadcastToClientsInTunnel(msg)
}
