package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("./test room server")
		return
	}

	room := os.Args[1]
	key := os.Args[2]

	statRoom(room, key)
}
