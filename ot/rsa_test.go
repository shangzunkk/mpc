//
// rsa_test.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package ot

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func benchmark(b *testing.B, keySize int) {
	l0, _ := NewLabel(rand.Reader)
	l1, _ := NewLabel(rand.Reader)

	sender, err := NewSender(keySize, map[int]Wire{
		0: Wire{
			L0: l0,
			L1: l1,
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	receiver, err := NewReceiver(sender.PublicKey())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sXfer, err := sender.NewTransfer(0)
		if err != nil {
			b.Fatal(err)
		}
		rXfer, err := receiver.NewTransfer(1)
		if err != nil {
			b.Fatal(err)
		}
		err = rXfer.ReceiveRandomMessages(sXfer.RandomMessages())
		if err != nil {
			b.Fatal(err)
		}

		sXfer.ReceiveV(rXfer.V())
		err = rXfer.ReceiveMessages(sXfer.Messages())
		if err != nil {
			b.Fatal(err)
		}

		m, bit := rXfer.Message()
		var ret int
		if bit == 0 {
			ret = bytes.Compare(l0.Bytes(), m)
		} else {
			ret = bytes.Compare(l1.Bytes(), m)
		}
		if ret != 0 {
			b.Fatal("Verify failed!\n")
		}
	}
}

func BenchmarkOT512(b *testing.B) {
	benchmark(b, 512)
}

func BenchmarkOT1024(b *testing.B) {
	benchmark(b, 1024)
}

func BenchmarkOT2048(b *testing.B) {
	benchmark(b, 2048)
}
