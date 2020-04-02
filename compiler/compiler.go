//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package compiler

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/markkurossi/mpc/circuit"
	"github.com/markkurossi/mpc/compiler/ast"
	"github.com/markkurossi/mpc/compiler/utils"
)

type Compiler struct {
	params   *utils.Params
	packages map[string]*ast.Package
}

func NewCompiler(params *utils.Params) *Compiler {
	return &Compiler{
		params:   params,
		packages: make(map[string]*ast.Package),
	}
}

func (c *Compiler) Compile(data string) (*circuit.Circuit, ast.Annotations,
	error) {
	return c.compile("{data}", strings.NewReader(data))
}

func (c *Compiler) CompileFile(file string) (*circuit.Circuit, ast.Annotations,
	error) {

	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return c.compile(file, f)
}

func (c *Compiler) compile(source string, in io.Reader) (
	*circuit.Circuit, ast.Annotations, error) {

	logger := utils.NewLogger(os.Stdout)
	pkg, err := c.parse(source, in, logger, ast.NewPackage("main"))
	if err != nil {
		return nil, nil, err
	}

	code, annotation, err := pkg.Compile(c.packages, logger, c.params)
	if err != nil {
		return nil, nil, err
	}
	if c.params.NoCircCompile {
		return nil, annotation, nil
	}
	circ, err := code.CompileCircuit(c.params)
	if err != nil {
		return nil, nil, err
	}
	return circ, annotation, nil
}

func (c *Compiler) parse(source string, in io.Reader, logger *utils.Logger,
	pkg *ast.Package) (*ast.Package, error) {

	parser := NewParser(source, c, logger, in)
	pkg, err := parser.Parse(pkg)
	if err != nil {
		return nil, err
	}
	c.packages[pkg.Name] = pkg

	for alias, name := range pkg.Imports {
		_, err := c.parsePkg(alias, name)
		if err != nil {
			return nil, err
		}
	}

	return pkg, nil
}

func (c *Compiler) parsePkg(alias, name string) (*ast.Package, error) {
	pkg, ok := c.packages[alias]
	if ok {
		return pkg, nil
	}
	pkg = ast.NewPackage(alias)

	if c.params.Verbose {
		fmt.Printf("looking for package %s (%s)\n", alias, name)
	}
	prefix := "go/src/github.com/markkurossi/mpc/pkg"

	dir := path.Join(os.Getenv("HOME"), prefix, name)
	df, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("package %s not found: %s", name, err)
	}
	defer df.Close()

	files, err := df.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("package %s not found: %s", name, err)
	}
	var mpcls []string
	for _, f := range files {
		if strings.HasSuffix(f, ".mpcl") {
			mpcls = append(mpcls, f)
		}
	}
	if len(mpcls) == 0 {
		return nil, fmt.Errorf("package %s is empty", name)
	}

	for _, mpcl := range mpcls {
		fp := path.Join(dir, mpcl)

		if c.params.Verbose {
			fmt.Printf(" - parsing %v\n", fp)
		}

		f, err := os.Open(fp)
		if err != nil {
			fmt.Printf("pkg not found: %s\n", err)
			return nil, fmt.Errorf("error reading package %s: %s", name, err)
		}
		defer f.Close()

		pkg, err = c.parse(fp, f, utils.NewLogger(os.Stdout), pkg)
		if err != nil {
			return nil, err
		}
	}
	return pkg, nil
}
