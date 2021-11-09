package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "golang.org:80")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(conn)
}
