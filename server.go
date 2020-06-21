package main

import (
	router "chatting-example/router"
)

func main() {
	e := router.Init()
	// Server
	e.Logger.Fatal(e.Start(":1213"))
}
