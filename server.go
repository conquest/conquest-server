package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type ClientMap struct {
	sync.RWMutex
	items map[int]*Client
}

func NewClientMap() *ClientMap {
	return &ClientMap{
		items: make(map[int]*Client),
	}
}

func (cmap *ClientMap) Get(key int) (c *Client, ok bool) {
	cmap.Lock()
	defer cmap.Unlock()

	c, ok = cmap.items[key]
	return
}

func (cmap *ClientMap) Set(key int, client *Client) {
	cmap.Lock()
	defer cmap.Unlock()

	cmap.items[key] = client
}

func (cmap *ClientMap) Delete(key int) {
	cmap.Lock()
	defer cmap.Unlock()

	delete(cmap.items, key)
}

func (cmap *ClientMap) Len() int {
	return len(cmap.items)
}

func (cmap *ClientMap) Iter() <-chan *Client {
	c := make(chan *Client)

	go func() {
		cmap.Lock()
		defer cmap.Unlock()

		for _, cl := range cmap.items {
			c <- cl
		}
		close(c)
	}()

	return c
}

type Packet struct {
	Block  Tile `json:"tile"`
	Player struct {
		Id    int    `json:"id"`
		Color string `json:"color"`
	} `json:"player"`
}

type Client struct {
	Conn    net.Conn
	Decoder *json.Decoder
	Pack    Packet
	Id      int
}

func message(msg interface{}) []byte {
	mar, _ := json.Marshal(msg)
	return append(mar, []byte("\n")...)
}

func (client *Client) Read() {
	for {
		var pack Packet
		err := client.Decoder.Decode(&pack)

		if err != nil {
			client.Conn.Close()
			clients.Delete(client.Id)

			if clients.Len() == 0 {
				game.Reset()
			}

			fmt.Println("Client disconnected with id:", client.Id)
			return
		} else {
			client.Pack = pack
			client.Pack.Player.Id = client.Id
			go game.Read(client.Pack.Block)
			go client.Write()
		}
	}
}

func (client *Client) Write() {
	data := message(client.Pack)
	go broadcast(data)
}

func (client *Client) sync() {
	us := message(client.Id)
	client.Conn.Write(us)

	for cl := range clients.Iter() {
		if cl != client {
			opps := message(cl.Pack.Player)
			client.Conn.Write(opps)
		}
	}

	latest := message(game)
	client.Conn.Write(latest)
}

func NewClient(conn net.Conn) *Client {
	decode := json.NewDecoder(conn)

	client := &Client{
		Conn:    conn,
		Decoder: decode,
		Id:      id,
	}
	id++
	client.sync()
	go client.Read()

	return client
}

func broadcast(msg []byte) {
	for client := range clients.Iter() {
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
			if clients.Len() == 0 {
				clock.Stop()
				fmt.Println("Clock has stopped")
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
			if clients.Len() == 0 {
				refresh.Stop()
				fmt.Println("Refreshing has stopped")
				return
			}

			fmt.Println("refresh")
			go broadcast([]byte("refresh\n"))
			game.Update()
		}
	}()
}

var (
	id      = 1
	game    *Map
	clients = NewClientMap()
)

func main() {
	file := flag.String("m", "maps/new-york.json", "The map that the server will run.")
	port := flag.String("p", "8080", "The port that the server will be open on.")
	flag.Parse()

	level, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Map selected:", *file)

	json.Unmarshal(level, &game)
	game.Initialize()

	fmt.Println("Starting server")

	clients = NewClientMap()
	l, err := net.Listen("tcp", ":"+*port)
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
		clients.Set(client.Id, client)
		fmt.Println("New client with id:", client.Id)

		if clients.Len() == 1 {
			startClock()
		}
	}
}
