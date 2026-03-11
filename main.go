package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bus
	bus := BuildBus(ctx)

	// send and receive via openai llm
	go Attach(bus, "openai", OpenAiLLM, ctx)

	// CLI
	Attach(bus, "CLI", Console, ctx)
	fmt.Printf("After console")
}
