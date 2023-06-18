package main

type Message struct {
	Sender  Address
	Text    string
	Counter int
}

type Envelope struct {
	Target Address
	Text   string
}
