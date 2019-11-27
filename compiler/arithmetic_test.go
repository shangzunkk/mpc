//
// arithmetic_test.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package compiler

import (
	"bufio"
	"io"
	"math/big"
	"testing"

	"github.com/markkurossi/mpc/circuit"
)

type Test struct {
	Name    string
	Operand string
	Bits    int
	Eval    func(a *big.Int, b *big.Int) *big.Int
	Code    string
}

var tests = []Test{
	Test{
		Name:    "Add",
		Operand: "+",
		Bits:    2,
		Eval: func(a *big.Int, b *big.Int) *big.Int {
			result := big.NewInt(0)
			result.Add(a, b)
			return result
		},
		Code: `
package main
func main(a, b int2) int3 {
    return a + b
}
`,
	},
	Test{
		Name:    "Multiply",
		Operand: "*",
		Bits:    2,
		Eval: func(a *big.Int, b *big.Int) *big.Int {
			result := big.NewInt(0)
			result.Mul(a, b)
			return result
		},
		Code: `
package main
func main(a, b int3) int6 {
    return a * b
}
`,
	},
}

func TestAdd(t *testing.T) {
	for _, test := range tests {
		circ, err := Compile(test.Code)
		if err != nil {
			t.Fatalf("Failed to compile test %s: %s", test.Name, err)
		}

		var key [32]byte

		limit := 1 << test.Bits

		for g := 0; g < limit; g++ {
			for e := 0; e < limit; e++ {
				gr, ew := io.Pipe()
				er, gw := io.Pipe()

				gio := bufio.NewReadWriter(
					bufio.NewReader(gr),
					bufio.NewWriter(gw))
				eio := bufio.NewReadWriter(
					bufio.NewReader(er),
					bufio.NewWriter(ew))

				gInput := big.NewInt(int64(g))
				eInput := big.NewInt(int64(e))

				go func() {
					_, err := circuit.Garbler(gio, circ, gInput, key[:], false)
					if err != nil {
						t.Fatalf("Garbler failed: %s\n", err)
					}
				}()

				result, err := circuit.Evaluator(eio, circ, eInput, key[:])
				if err != nil {
					t.Fatalf("Evaluator failed: %s\n", err)
				}

				expected := test.Eval(gInput, eInput)

				if expected.Cmp(result) != 0 {
					t.Errorf("%s failed: %s %s %s = %s, expected %s\n",
						test.Name, gInput, test.Operand, eInput, result,
						expected)
				}
			}
		}
	}
}
