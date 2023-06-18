package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

var self Address

type Address uint32

var reader *bufio.Scanner

var counter = 10

// main receives the actor identifier as an argument.
func main() {
	atoi, _ := strconv.Atoi(os.Args[1])
	self = Address(atoi)
	reader = bufio.NewScanner(os.Stdin)
}

// receive reads from buf[0:size] for the given size,
// decodes the given message, does its own business
// and then writes to buf
// 0 or more messages.
//
//export receive
func receive() {
	if counter == 0 {
		log.Printf("This actor is dead.")
		return
	}

	reader.Scan()
	message := Message{}
	err := json.Unmarshal(reader.Bytes(), &message)
	if err != nil {
		panic(err)
	}

	// We received a message, print out a message.
	log.Printf("Received message from %d: '%s'\n", message.Sender, message.Text)

	// This is a ping-pong; we reply to the sender with another message.
	message.Sender.Tell(
		Message{
			Sender: self,
			Text:   fmt.Sprintf("ping from %d", self)})

	counter--
}

// Tell writes a message to buf and increases the outgoing count.
// The outgoing count is always at offset 0. When the actor returns
// the host reads the count and then collects and dispatches all the messages
// to every address.
func (a Address) Tell(message Message) {
	bytes, err := json.Marshal(message)
	envelope := Envelope{
		Target: a,
		Text:   string(bytes),
	}
	bytes, err = json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(append(bytes, '\n'))
}
