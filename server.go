package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Packet struct {
	Tiles []struct {
		Index  uint16 `json:"index"`
		Owner  uint8  `json:"owner"`
		Troops uint32 `json:"troops"`
	} `json:"tiles"`
	Troops []struct {
		Owner    uint8  `json:"owner"`
		Quantity uint32 `json:"quantity"`
	} `json:"troops"`
}

// TODO multiple clients
func main() {
	fmt.Println("starting server")

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("error connecting: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting: ", err.Error())
			os.Exit(1)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	for {
		decoder := json.NewDecoder(conn)

		var pack Packet
		err := decoder.Decode(&pack)

		if err != nil {
			conn.Close()
			return
		} else {
			fmt.Println("packet: ", pack)
		}
	}
}
