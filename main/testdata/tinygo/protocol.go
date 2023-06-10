package main

import "encoding/binary"

type Message struct {
	sender  Address
	message string
}

func (m *Message) decode(buff []byte) {
	m.sender = Address(binary.LittleEndian.Uint32(buff[:4]))
	m.message = string(buff[4:])
}

func (m *Message) encode(target Address, buf []byte) uint32 {
	// target: uint32
	binary.LittleEndian.PutUint32(buf[0:4], uint32(target))
	// total length is 4 bytes + len(message)
	// on decode we infer len(message) from the total size - 4
	mlen := uint32(4 + len(m.message))
	binary.LittleEndian.PutUint32(buf[4:8], mlen)
	// message body
	binary.LittleEndian.PutUint32(buf[8:12], uint32(m.sender))
	sz := 12 + uint32(len(m.message))
	copy(buf[12:sz], m.message)
	return sz
}
