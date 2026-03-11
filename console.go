package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

// Console output
func Console(inbox MessageReceiver, outbox MessageSender, _ Done, ctx context.Context) {

	input := make(chan string)
	go handleConsoleInput(input)

	for {
		select {
		case <-ctx.Done():
			return
		case received := <-inbox:
			if received.Response != nil {
				fmt.Printf("%s\n", *received.Response)
			}
		case stdin := <-input:
			if stdin == "q" {
				return
			}
			outbox <- Message{
				ID:   uuid.NewString(),
				Call: stdin,
			}
		}
	}

}

func handleConsoleInput(input chan string) {

	reader := bufio.NewReader(os.Stdin)
	for {
		value, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Stdin error: %s", err)
			continue
		}
		input <- strings.TrimSuffix(value, "\n")
	}
}
