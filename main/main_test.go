package main

import (
	"context"
	_ "embed"
	"encoding/binary"
	"testing"
)

//go:embed "testdata/tinygo/pingpong.wasm"
var pingpong []byte

func TestActorSystem_ActorOf(t *testing.T) {
	system := NewActorSystem(context.Background())
	ping := system.ActorOf("pinger", pingpong)
	pong := system.ActorOf("ponger", pingpong)

	var m []byte
	m = binary.LittleEndian.AppendUint32(m, ping.Address())
	m = append(m, []byte("ping")...)

	ping.Tell(Message{pong.Address(), m})
	system.Wait()
}
