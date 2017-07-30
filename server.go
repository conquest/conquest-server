package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Tile struct {
	Index  uint16 `json:"index"`
	Owner  uint8  `json:"owner"`
	Troops uint32 `json:"troops"`
}

type Packet struct {
	Tiles  []Tile `json:"tiles"`
	Player struct {
		Id      uint8  `json:"id"`
		Color   string `json:"color"`
		Capital Tile   `json:"capital"`
	} `json:"player"`
}

func (pack Packet) IsEmpty() bool {
	return pack.Tiles == nil
}

var clients map[int]*Client

type Client struct {
	Conn    net.Conn
	Decoder *json.Decoder
	Pack    Packet
	Running bool
	Id      uint8
}

func (client *Client) Read() {
	for {
		var pack Packet
		err := client.Decoder.Decode(&pack)

		if err != nil {
			client.Conn.Close()
			client.Running = false
			delete(clients, int(client.Id)-1)
			fmt.Println("Client disconnected with id:", client.Id)
			return
		} else {
			client.Pack = pack
			client.Pack.Player.Id = client.Id
		}
	}
}

func (client *Client) Write() {
	for {
		if !client.Running {
			return
		}
		for _, cl := range clients {
			if !cl.Pack.IsEmpty() {
				var data []byte
				if cl == client {
					data, _ = json.Marshal(cl.Pack.Player)
				} else {
					data, _ = json.Marshal(cl.Pack)
				}
				data = append(data, []byte("\n")...)
				client.Conn.Write(data)
			}
		}
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func NewClient(conn net.Conn) *Client {
	decode := json.NewDecoder(conn)

	client := &Client{
		Conn:    conn,
		Decoder: decode,
		Running: true,
	}
	client.Listen()

	return client
}

func main() {
	fmt.Println("Starting server")

	clients = make(map[int]*Client)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		client := NewClient(conn)
		clients[len(clients)] = client
		client.Id = uint8(len(clients))
		fmt.Println("New client with id:", client.Id)
	}
}
