//
// codegen.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package circuits

import (
	"github.com/markkurossi/mpc/circuit"
)

type Compiler struct {
	N1         circuit.IO
	N2         circuit.IO
	N3         circuit.IO
	Inputs     []*Wire
	Outputs    []*Wire
	Gates      []Gate
	nextWireID uint32
	nextGateID uint32
	pending    []Gate
	assigned   []Gate
	compiled   []*circuit.Gate
	zeroWire   *Wire
}

func NewIO(size int) circuit.IO {
	return circuit.IO{circuit.IOArg{Size: size}}
}

func NewCompiler(n1, n2, n3 circuit.IO) *Compiler {
	result := &Compiler{
		N1: n1,
		N2: n2,
		N3: n3,
	}

	for i := 0; i < n1.Size()+n2.Size(); i++ {
		result.Inputs = append(result.Inputs, NewWire())
	}
	for i := 0; i < n3.Size(); i++ {
		result.Outputs = append(result.Outputs, NewOutputWire())
	}

	return result
}

func (c *Compiler) ZeroWire() *Wire {
	if c.zeroWire == nil {
		c.zeroWire = NewWire()
		var head []*Wire
		head = append(head, c.Inputs[0:c.N1.Size()]...)
		head = append(head, c.zeroWire)
		head = append(head, c.Inputs[c.N1.Size():]...)
		c.Inputs = head
		c.N1 = append(c.N1, circuit.IOArg{
			Name: "0",
			Size: 1,
		})
	}
	return c.zeroWire
}

func (c *Compiler) Zero(o *Wire) {
	w := NewWire()
	c.AddGate(NewINV(c.ZeroWire(), w))
	c.AddGate(NewINV(w, o))
}

func (c *Compiler) One(o *Wire) {
	c.AddGate(NewINV(c.ZeroWire(), o))
}

func (c *Compiler) AddGate(gate Gate) {
	c.Gates = append(c.Gates, gate)
}

func (c *Compiler) NextWireID() uint32 {
	ret := c.nextWireID
	c.nextWireID++
	return ret
}

func (c *Compiler) NextGateID() uint32 {
	ret := c.nextGateID
	c.nextGateID++
	return ret
}

func (c *Compiler) Compile() *circuit.Circuit {
	for _, w := range c.Inputs {
		w.Assign(c)
	}
	for len(c.pending) > 0 {
		gate := c.pending[0]
		c.pending = c.pending[1:]
		gate.Assign(c)
	}
	// Assign outputs.
	for _, w := range c.Outputs {
		if w.Assigned {
			panic("Output already assigned")
		}
		w.ID = c.NextWireID()
		w.Assigned = true
	}

	// Compile circuit.
	for _, gate := range c.assigned {
		gate.Compile(c)
	}

	result := &circuit.Circuit{
		NumGates: int(c.nextGateID),
		NumWires: int(c.nextWireID),
		N1:       c.N1,
		N2:       c.N2,
		N3:       c.N3,
		Gates:    make(map[int]*circuit.Gate),
	}

	for _, gate := range c.compiled {
		result.Gates[int(gate.ID)] = gate
	}

	return result
}

type Wire struct {
	Output   bool
	Assigned bool
	ID       uint32
	Input    Gate
	Outputs  []Gate
}

func NewWire() *Wire {
	return &Wire{}
}

func NewOutputWire() *Wire {
	return &Wire{
		Output: true,
	}
}

func (w *Wire) Assign(c *Compiler) {
	if w.Output {
		return
	}
	if !w.Assigned {
		w.ID = c.NextWireID()
		w.Assigned = true
	}
	for _, output := range w.Outputs {
		output.Visit(c)
	}
}

func (w *Wire) SetInput(gate Gate) {
	if w.Input != nil {
		panic("Input gate already set")
	}
	w.Input = gate
}

func (w *Wire) AddOutput(gate Gate) {
	w.Outputs = append(w.Outputs, gate)
}

type Gate interface {
	String() string
	Visit(c *Compiler)
	Assign(c *Compiler)
	Compile(c *Compiler)
}
