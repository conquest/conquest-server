package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"
)

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

			if len(clients) == 0 {
				game.Reset()
			}

			fmt.Println("Client disconnected with id:", client.Id)
			return
		} else {
			client.Pack = pack
			client.Pack.Player.Id = client.Id
		}
	}
}

func (client *Client) Write() {
	for client.Running {
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

func (client *Client) Sync() {
	latest, _ := json.Marshal(game)
	latest = append(latest, []byte("\n")...)
	client.Conn.Write(latest)
}

func NewClient(conn net.Conn) *Client {
	decode := json.NewDecoder(conn)

	client := &Client{
		Conn:    conn,
		Decoder: decode,
		Running: true,
	}
	client.Sync()
	client.Listen()

	return client
}

func broadcast(msg []byte) {
	for _, client := range clients {
		client.Conn.Write(msg)
	}
}

func clockTime(time int) (ct string) {
	ct = ""
	if time/60 < 10 {
		ct += "0"
	}
	ct += fmt.Sprintf("%d:", time/60)
	if time%60 < 10 {
		ct += "0"
	}
	ct += fmt.Sprintf("%d", time%60)

	return
}

func startClock() {
	clock := time.NewTicker(time.Second)
	refresh := time.NewTicker(10 * time.Second)

	go func() {
		time := 0
		for range clock.C {
			if len(clients) == 0 {
				clock.Stop()
				return
			}

			time++
			msg := clockTime(time) + "\n"
			fmt.Print(msg)
			go broadcast([]byte(msg))
		}
	}()

	go func() {
		for range refresh.C {
			if len(clients) == 0 {
				refresh.Stop()
				return
			}

			fmt.Println("refresh")
			go broadcast([]byte("refresh\n"))
			game.Update()
		}
	}()
}

var game *Map

func main() {
	file := flag.String("m", "maps/new-york.json", "The map that the server will run.")
	flag.Parse()

	level, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Map selected:", *file)

	json.Unmarshal(level, &game)
	game.Initialize()

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

		if len(clients) == 1 {
			startClock()
		}
	}
}
