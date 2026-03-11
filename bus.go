package main

import (
	"context"
	"fmt"
)

// a message is something with an id, a "call" and, when provided a "response" to the call
type Message struct {
	ID       string
	Call     string
	Response *string
}

// an actor id is used to identify the actor listening to, or putting a message on
// to the bus (or back on to the bus if it is providing a response)
type ActorID string

// an actor message extends a message with the id of the actor
type ActorMessage struct {
	Message
	Actor ActorID
}

// Message receiver and sender are ways of getting messages to/from the bus
type MessageReceiver <-chan Message
type MessageSender chan<- Message

// Done lets an actor detach themselves from the bus
type Done chan<- struct{}

// if I want to register to receive messages, I send a ListenerPort message
type ListenerPort struct {
	Actor ActorID
	Port  MessageSender
}

// the bus has 3 channels - register, deregister and send (to send messages)
type Bus struct {
	register   chan<- ListenerPort
	deregister chan<- ActorID
	send       chan<- ActorMessage
}

// an Actor is a function which takes sender and receiver channels, and an extra
// for deregistering. These channels  are intercepted so that an Actor ID can be
// automatically managed
type Actor func(inbox MessageReceiver, outbox MessageSender, done Done, ctx context.Context)

// sets up the channels, starts a goroutine to listen for the three channels
func BuildBus(ctx context.Context) Bus {
	registerChannel := make(chan ListenerPort)
	deregisterChannel := make(chan ActorID)
	receiveChannel := make(chan ActorMessage)
	registered := make(map[ActorID]ListenerPort)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case registration := <-registerChannel:
				registered[registration.Actor] = registration
				fmt.Printf("[BUS] Registered: %s\n", registration.Actor)
			case idToDeregister := <-deregisterChannel:
				delete(registered, idToDeregister)
				fmt.Printf("[BUS] Deregistered: %s\n", idToDeregister)
			case message := <-receiveChannel:
				if message.Response == nil {
					fmt.Printf("[%s -> BUS] %s | Call: %s\n", message.Actor, message.ID, message.Call)
				} else {
					fmt.Printf("[%s -> BUS] %s | Call: %s | Response: %s\n", message.Actor, message.ID, message.Call, *message.Response)
				}
				for actor, r := range registered {
					if actor != message.Actor {
						r.Port <- message.Message
					}
				}
			}
		}
	}()
	return Bus{
		register:   registerChannel,
		deregister: deregisterChannel,
		send:       receiveChannel,
	}
}

func Attach(bus Bus, actorID ActorID, actor Actor, ctx context.Context) {
	// this will take messages to the actor
	toActor := make(chan Message)
	// this will get messages from the actor
	fromActor := make(chan Message)
	// the actor can signal they want to detach
	done := make(chan struct{})
	// tell the bus we have a new actor and give it the actor's message port
	bus.register <- ListenerPort{actorID, toActor}
	// start the actor with the three channels
	go actor(toActor, fromActor, done, ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-fromActor:
			// actor is sending a message, decorate and forward to the bus
			select {
			case bus.send <- ActorMessage{m, actorID}:
			case <-ctx.Done():
				return
			}
		case <-done:
			// actor wants to detach so deregister it
			select {
			case bus.deregister <- actorID:
			case <-ctx.Done():
			}
			return
		}
	}
}
