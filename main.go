package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	StartUDPServer()

	conn, err := ConnectUDP()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	StartClientListener(conn)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Message to send: ")

		// Scan() blocks until the user presses Enter
		if scanner.Scan() {
			input := scanner.Text() // Retrieve the input string
			SendUDP(input, conn)
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("[SCANNER] error:", err)
		}
	}
}
