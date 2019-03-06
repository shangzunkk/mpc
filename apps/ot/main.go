//
// main.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/markkurossi/mpc/ot"
)

func main() {
	sender, err := ot.NewSender(2048)
	if err != nil {
		log.Fatal(err)
	}

	receiver, err := ot.NewReceiver()
	if err != nil {
		log.Fatal(err)
	}

	receiver.ReceivePublicKey(sender.PublicKey())
	err = receiver.ReceiveRandomMessages(sender.RandomMessages())
	if err != nil {
		log.Fatal(err)
	}

	sender.ReceiveV(receiver.V())
	err = receiver.ReceiveMessages(sender.Messages())
	if err != nil {
		log.Fatal(err)
	}

	m, bit := receiver.Message()
	var ret int
	if bit == 0 {
		ret = bytes.Compare(sender.M0(), m)
	} else {
		ret = bytes.Compare(sender.M1(), m)
	}
	if ret != 0 {
		fmt.Printf("Verify failed!\n")
		os.Exit(1)
	}
}
