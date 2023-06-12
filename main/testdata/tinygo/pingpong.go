package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

const SHARED_BUF_SIZE = 1024

var buf [SHARED_BUF_SIZE]uint8

var self Address
var outgoing uint32 = 0
var currOffset uint32 = 0

type Address uint32

func main() {
	// main is necessary for tinygo
	reset()
}

// startup initializes the actor: receives its own address
// to make the actor self-aware.
//
// It also exports the address of the buf buffer
// to the host, so that we can share messages with it.
//
//export startup
func startup(addr uint32) *[SHARED_BUF_SIZE]uint8 {
	self = Address(addr)
	return &buf
}

// receive reads from buf[0:size] for the given size,
// decodes the given message, does its own business
// and then writes to buf
// 0 or more messages.
//
//export receive
func receive(size uint32) {
	reset()
	message := Message{}
	err := json.Unmarshal(buf[0:size], &message)
	if err != nil {
		panic(err)
	}

	// We received a message, print out a message.
	fmt.Printf("Received message from %d: '%s'\n", message.Sender, message.Text)

	// This is a ping-pong; we reply to the sender with another message.
	message.Sender.Tell(Message{Sender: self, Text: fmt.Sprintf("ping from %d", self)})
	done()
}

// Tell writes a message to buf and increases the outgoing count.
// The outgoing count is always at offset 0. When the actor returns
// the host reads the count and then collects and dispatches all the messages
// to every address.
func (a Address) Tell(message Message) {
	// increase the outgoing count
	outgoing++
	// update the header with the new count
	binary.LittleEndian.PutUint32(buf[:4], outgoing)

	// append another message to the buffer
	outBuf := buf[currOffset:]
	// target: uint32
	binary.LittleEndian.PutUint32(outBuf[:4], uint32(a))
	bytes, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(outBuf[4:8], uint32(len(bytes)))
	copy(outBuf[8:], bytes)
	currOffset += 4 + 4 + uint32(len(bytes))
}

func reset() {
	currOffset = 4
	outgoing = 0
}

func done() {
	// if the actor did not send out any message, write a 0 count header
	if outgoing == 0 {
		binary.LittleEndian.PutUint32(buf[:4], outgoing)
	}
}
