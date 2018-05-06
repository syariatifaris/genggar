package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var wg sync.WaitGroup

func main() {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", "127.0.0.1:1234")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing", err.Error())
		}

		fmt.Println("closing connection")
	}()

	fmt.Fprintf(conn, fmt.Sprint("[REG]:", conn.LocalAddr()))

	wg.Add(1)
	go func() {
		for {
			n, err := bufio.NewReader(conn).Read(p)
			if err == nil {
				fmt.Printf("[Server Says]: %s\n", string(p[0:n]))
			} else {
				fmt.Printf("Some error %v\n", err)
			}
		}
	}()

	wg.Wait()
}
