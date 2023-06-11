package main

import (
	"context"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"os"
	"sync"
	"sync/atomic"
)

func main() {
}

type Address uint32

type ActorRef interface {
	Tell(m Envelope)
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

type Envelope struct {
	sender Address
	body   []byte
}

type Actor struct {
	system *ActorSystem
	mod    api.Module
	in     chan Envelope
	recv   api.Function
	ptr    uint32
	addr   Address
}

func (s *ActorSystem) ActorOf(name string, bytes []byte) ActorRef {
	a := &Actor{}
	a.system = s
	a.in = make(chan Envelope, 32)
	ctx := context.Background()
	cfg := wazero.NewModuleConfig().WithStderr(os.Stderr).WithStdout(os.Stdout)
	mod, err := s.rt.InstantiateWithConfig(ctx, bytes, cfg)
	if err != nil {
		panic(err)
	}
	a.mod = mod
	a.recv = a.mod.ExportedFunction("receive")
	startup := a.mod.ExportedFunction("startup")
	a.addr = Address(s.gen.Add(1))
	s.wg.Add(1)
	s.actors[a.addr] = a
	println("created actor", a.addr)
	ptr, err := startup.Call(ctx, uint64(a.addr))
	if err != nil {
		panic(err)
	}
	a.ptr = uint32(ptr[0])
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
		// Allocate enough space for m.body.
		sz := uint32(len(m.body))

		// Write to the shared buffer the message body.
		a.mod.Memory().Write(a.ptr, m.body)

		// Invoke the actor receive.
		_, _ = a.recv.Call(ctx, uint64(sz))
		// The actor appends in the same buffer, overwriting it,
		// a sequence of outgoing messages, prefixed by the total count.

		// Now read from the shared buffer.
		off := a.ptr
		count, _ := a.mod.Memory().ReadUint32Le(off)
		off += 4

		// For each message in the buffer
		for i := uint32(0); i < count; i++ {
			// Prefix 4-byte address, 4-byte size, then contents.
			address, _ := a.mod.Memory().ReadUint32Le(off)
			off += 4
			sz, _ = a.mod.Memory().ReadUint32Le(off)
			off += 4
			bytes, _ := a.mod.Memory().Read(off, sz)
			off += sz
			m = Envelope{
				sender: Address(address),
				body:   bytes,
			}
			// We have decoded the header of the message
			// We don't need to decode the contents, we just
			// pass-through them to the other actors.
			a.system.actors[Address(address)].in <- m
		}

	}
}

func (a *Actor) ActorRef() ActorRef {
	return a
}

func (a *Actor) Tell(m Envelope) {
	println("tell message to", m.sender)
	actor, ok := a.system.actors[m.sender]
	if !ok {
		println("no such actor", m.sender)
	}
	actor.in <- m
}

func (a *Actor) Address() Address {
	return a.addr
}
