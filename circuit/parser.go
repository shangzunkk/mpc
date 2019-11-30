//
// parser.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package circuit

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strconv"
)

type Operation byte

const (
	XOR Operation = iota
	AND
	OR
	INV
)

var reParts = regexp.MustCompilePOSIX("[[:space:]]+")

func (op Operation) String() string {
	switch op {
	case XOR:
		return "XOR"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case INV:
		return "INV"
	default:
		return fmt.Sprintf("{Operation %d}", op)
	}
}

type IOArg struct {
	Name string
	Type string
	Size int
}

type IO []IOArg

func (io IO) Size() int {
	var sum int
	for _, a := range io {
		sum += a.Size
	}
	return sum
}

func (io IO) Parse(inputs []string) ([]*big.Int, error) {
	var result []*big.Int

	for idx, _ := range io {
		i := new(big.Int)
		// XXX Type checks
		_, ok := i.SetString(inputs[idx], 10)
		if !ok {
			return nil, fmt.Errorf("Invalid input: %s", inputs[idx])
		}
		result = append(result, i)
	}
	return result, nil
}

func (io IO) String() string {
	var str = ""
	for i, a := range io {
		if i > 0 {
			str += ", "
		}
		if len(a.Name) > 0 {
			str += a.Name + ":"
		}
		str += a.Type
	}
	return str
}

func (io IO) Split(in *big.Int) []*big.Int {
	var result []*big.Int
	var bit int
	for _, arg := range io {
		r := big.NewInt(0)
		for i := 0; i < arg.Size; i++ {
			if in.Bit(bit) == 1 {
				r = big.NewInt(0).SetBit(r, i, 1)
			}
			bit++
		}
		result = append(result, r)
	}
	return result
}

type Circuit struct {
	NumGates int
	NumWires int
	N1       IO
	N2       IO
	N3       IO
	Gates    map[int]*Gate
}

func (c *Circuit) String() string {
	return fmt.Sprintf("#gates=%d, #wires=%d n1=%d, n2=%d, n3=%d",
		c.NumGates, c.NumWires, c.N1.Size(), c.N2.Size(), c.N3.Size())
}

func (c *Circuit) Marshal(out io.Writer) {
	fmt.Fprintf(out, "%d %d\n", c.NumGates, c.NumWires)
	fmt.Fprintf(out, "%d %d %d\n", c.N1.Size(), c.N2.Size(), c.N3.Size())
	fmt.Fprintf(out, "\n")

	type kv struct {
		Key   uint32
		Value *Gate
	}
	var gates []kv

	for _, gate := range c.Gates {
		gates = append(gates, kv{
			Key:   gate.ID,
			Value: gate,
		})
	}
	sort.Slice(gates, func(i, j int) bool {
		return gates[i].Key < gates[j].Key
	})

	for _, gate := range gates {
		g := gate.Value
		fmt.Fprintf(out, "%d %d", len(g.Inputs), len(g.Outputs))
		for _, w := range g.Inputs {
			fmt.Fprintf(out, " %d", w)
		}
		for _, w := range g.Outputs {
			fmt.Fprintf(out, " %d", w)
		}
		fmt.Fprintf(out, " %s\n", g.Op)
	}
}

type Gate struct {
	ID      uint32
	Inputs  []Wire
	Outputs []Wire
	Op      Operation
}

type Wire uint32

func (w Wire) ID() int {
	return int(w)
}

func (w Wire) String() string {
	return fmt.Sprintf("w%d", w)
}

func Parse(in io.Reader) (*Circuit, error) {
	r := bufio.NewReader(in)

	// NumGates NumWires
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	if len(line) != 2 {
		fmt.Printf("Line: %v\n", line)
		return nil, errors.New("Invalid 1st line")
	}
	numGates, err := strconv.Atoi(line[0])
	if err != nil {
		return nil, err
	}
	numWires, err := strconv.Atoi(line[1])
	if err != nil {
		return nil, err
	}

	// N1 N2 N3
	line, err = readLine(r)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if len(line) != 3 {
		return nil, errors.New("Invalid 2nd line")
	}
	n1, err := strconv.Atoi(line[0])
	if err != nil {
		return nil, err
	}
	n2, err := strconv.Atoi(line[1])
	if err != nil {
		return nil, err
	}
	n3, err := strconv.Atoi(line[2])
	if err != nil {
		return nil, err
	}

	gates := make(map[int]*Gate)
	for gate := 0; ; gate++ {
		line, err = readLine(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(line) < 3 {
			return nil, fmt.Errorf("Invalid gate: %v", line)
		}
		n1, err := strconv.Atoi(line[0])
		if err != nil {
			return nil, err
		}
		n2, err := strconv.Atoi(line[1])
		if err != nil {
			return nil, err
		}
		if 2+n1+n2+1 != len(line) {
			return nil, fmt.Errorf("Invalid gate: %v", line)
		}

		var inputs []Wire
		for i := 0; i < n1; i++ {
			v, err := strconv.Atoi(line[2+i])
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, Wire(v))
		}

		var outputs []Wire
		for i := 0; i < n2; i++ {
			v, err := strconv.Atoi(line[2+n1+i])
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, Wire(v))
		}
		var op Operation
		switch line[len(line)-1] {
		case "XOR":
			op = XOR
		case "AND":
			op = AND
		case "OR":
			op = OR
		case "INV":
			op = INV
		default:
			return nil, fmt.Errorf("Invalid operation '%s'", line[len(line)-1])
		}

		gates[gate] = &Gate{
			ID:      uint32(gate),
			Inputs:  inputs,
			Outputs: outputs,
			Op:      op,
		}
	}

	return &Circuit{
		NumGates: numGates,
		NumWires: numWires,
		N1:       []IOArg{IOArg{Size: n1}},
		N2:       []IOArg{IOArg{Size: n2}},
		N3:       []IOArg{IOArg{Size: n3}},
		Gates:    gates,
	}, nil
}

func readLine(r *bufio.Reader) ([]string, error) {
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(line) == 1 {
			continue
		}
		parts := reParts.Split(line[:len(line)-1], -1)
		if len(parts) > 0 {
			return parts, nil
		}
	}
}
