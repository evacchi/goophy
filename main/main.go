package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

func main() {
}

type Address uint32

type ActorRef interface {
	Tell(m EncodedMessage)
	Address() Address
}

type ActorSystem struct {
	gen    atomic.Uint32
	rt     wazero.Runtime
	actors map[Address]*Actor
	wg     sync.WaitGroup
}

func NewActorSystem(ctx context.Context) *ActorSystem {
	rt := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	return &ActorSystem{rt: rt, actors: map[Address]*Actor{}}
}

type EncodedMessage []byte

type Actor struct {
	system   *ActorSystem
	mod      api.Module
	in       chan EncodedMessage
	recv     api.Function
	ptr      uint32
	addr     Address
	actorIn  io.Writer
	actorOut *bufio.Scanner
}

func (s *ActorSystem) ActorOf(name string, buf []byte) ActorRef {
	a := &Actor{}
	a.system = s
	a.in = make(chan EncodedMessage, 32)
	var stdin, stdout bytes.Buffer

	a.actorIn = &stdin
	a.actorOut = bufio.NewScanner(&stdout)
	a.addr = Address(s.gen.Add(1))
	ctx := context.Background()
	cfg := wazero.NewModuleConfig().
		WithStdin(&stdin).
		WithStderr(os.Stderr).
		WithStdout(&stdout).
		WithArgs("actor", strconv.Itoa(int(a.addr)))
	mod, err := s.rt.InstantiateWithConfig(ctx, buf, cfg)
	if err != nil {
		panic(err)
	}
	a.mod = mod
	a.recv = a.mod.ExportedFunction("receive")
	s.wg.Add(1)
	s.actors[a.addr] = a
	println("created actor", a.addr)
	if err != nil {
		panic(err)
	}
	go a.receive()
	return a.ActorRef()
}

func (s *ActorSystem) Wait() {
	s.wg.Wait()
}

func (a *Actor) receive() {
	ctx := context.Background()
	for {
		m := <-a.in
		// Write to the shared buffer the message body.
		a.actorIn.Write(m)
		a.actorIn.Write([]byte{'\n'})

		// Invoke the actor receive.
		_, _ = a.recv.Call(ctx)

		a.actorOut.Scan()
		bytes := a.actorOut.Bytes()
		e := Envelope{}
		json.Unmarshal(bytes, &e)
		a.system.actors[e.Target].Tell(EncodedMessage(e.Text))
	}
}

type Envelope struct {
	Target Address
	Text   string
}

func (a *Actor) ActorRef() ActorRef {
	return a
}

func (a *Actor) Tell(m EncodedMessage) {
	a.in <- m
}

func (a *Actor) Address() Address {
	return a.addr
}
