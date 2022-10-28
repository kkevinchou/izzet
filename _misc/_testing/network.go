package main

import (
	"time"

	"github.com/kkevinchou/izzet/lib/network"
)

func main() {
	host := "localhost"
	port := "8080"
	connectionType := "tcp"

	server := network.NewServer(host, port, connectionType, 18)
	err := server.Start()
	if err != nil {
		panic(err)
	}

	client, id, err := network.Connect(host, port, connectionType)
	if err != nil {
		panic(err)
	}

	if id == network.UnsetClientID {
		panic(err)
	}

	err = client.SendMessage(network.MessageTypeInput, nil)
	if err != nil {
		panic(err)
	}

	time.Sleep(10000 * time.Second)
}
