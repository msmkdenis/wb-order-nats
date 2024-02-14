package main

import (
	"github.com/msmkdenis/wb-order-nats/internal/app/fakeproducer"
)

func main() {
	fakeproducer.Run("test-cluster", "test-client-sender", "http://127.0.0.1:4222/", 100)
}
