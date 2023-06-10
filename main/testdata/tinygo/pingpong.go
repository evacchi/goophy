package main

import (
	"encoding/binary"
	"fmt"
)

const SHARED_BUF_SIZE = 1024

var buf [SHARED_BUF_SIZE]uint8

var self Address
var outgoing uint32 = 0
var currOffset uint32 = 0

type Address uint32

func main() {
	println("hello")
}

//export startup
func startup(addr uint32) *[SHARED_BUF_SIZE]uint8 {
	self = Address(addr)
	return &buf
}

//export receive
func receive(size uint32) {
	reset()
	message := Message{}
	message.decode(buf[0:size])
	fmt.Printf("Received message from %d: '%s'\n", message.sender, message.message)
	message.sender.Tell(Message{sender: self, message: fmt.Sprintf("ping from %d", self)})
}

func (a Address) Tell(message Message) {
	outgoing++
	// buffer header
	binary.LittleEndian.PutUint32(buf[:4], outgoing)
	outBuf := buf[currOffset:]
	sz := message.encode(a, outBuf)
	currOffset += sz
}

func reset() {
	currOffset = 4
	outgoing = 0
}
