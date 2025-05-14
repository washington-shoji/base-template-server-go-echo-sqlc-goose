package main

import "go-echo-server-template/server"

func main() {
	if err := server.Start(); err != nil {
		panic(err)
	}
}
