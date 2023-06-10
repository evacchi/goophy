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

type Message struct {
	sender  Address
	message string
}

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
	message := read(size)
	fmt.Printf("Received message from %d: '%s'\n", message.sender, message.message)
	message.sender.Tell(Message{sender: self, message: fmt.Sprintf("ping from %d", self)})
}

func (a Address) Tell(message Message) {
	outgoing++
	// buffer header
	binary.LittleEndian.PutUint32(buf[:4], outgoing)

	// message header: target + total len
	buff := buf[4:]

	binary.LittleEndian.PutUint32(buff[0:4], uint32(a))
	mlen := uint32(4 + 4 + len(message.message))
	binary.LittleEndian.PutUint32(buff[4:8], mlen)
	// message body
	binary.LittleEndian.PutUint32(buff[8:12], uint32(message.sender))
	copy(buff[12:12+uint32(len(message.message))], message.message)
	
}

func read(size uint32) Message {
	currOffset = 4
	outgoing = 0
	buff := buf[0:size]
	address := Address(binary.LittleEndian.Uint32(buff[:4]))
	message := string(buff[4 : size-4])
	m := Message{
		sender:  address,
		message: message,
	}
	return m
}
