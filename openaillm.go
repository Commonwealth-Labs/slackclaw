package main

import (
	"context"
	"fmt"
	"time"
)

func OpenAiLLM(inbox MessageReceiver, outbox MessageSender, _ Done, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case received := <-inbox:
		fmt.Printf("LLM is thinking...\n")
		time.Sleep(time.Second * 5)
		response := "The answer is 42"
		received.Response = &response
		outbox <- received
	}
}
