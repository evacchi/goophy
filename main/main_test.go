package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed "testdata/tinygo/pingpong.wasm"
var pingpong []byte

func TestActorSystem_ActorOf(t *testing.T) {

	system := NewActorSystem(context.Background())
	ping := system.ActorOf("pinger", pingpong)
	pong := system.ActorOf("ponger", pingpong)

	//var m []byte
	//m = binary.LittleEndian.AppendUint32(m, uint32(ping.Address()))
	//m = append(m, []byte("ping")...)

	s := struct {
		Sender Address
		Text   string
	}{
		Sender: pong.Address(),
		Text:   "Begin",
	}

	m, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	ping.Tell(m)
	system.Wait()
}
