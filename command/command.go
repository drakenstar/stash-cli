package command

import (
	"errors"
	"fmt"
	"io"
)

// Config is a map of command names to configurations.  Each entry in this map defines a command that can be resolved
// by passing an Iterator to Resolve().  Each Command can also have sub-commands, or a Resolve func that will be called
// once it's determined that the input command is matched with this config.
type Config map[string]struct {
	Resolve     func(Iterator) (any, error)
	SubCommands Config
}

// TODO not many of these errors are useful in themselves - we probably want most of these errors to be displayed to
// users because it's their input that is directly leading to them, and they can take action to rectify.
var (
	ErrNoInput    = errors.New("no arguments returned")
	ErrNoCommand  = errors.New("no matched command")
	ErrNoResolver = errors.New("no resolver configured")
)

// Resolve takes an Iterator and calls Next until it can match a Command.  Once a match is made, it will call the
// command Resolve function and pass it the remaining arguments.
func (s Config) Resolve(it Iterator) (any, error) {
	if s == nil {
		return nil, fmt.Errorf("nil command config passed")
	}
	p := &peekIter{it: it}

	arg, err := p.Next()
	if err == io.EOF {
		return nil, ErrNoInput
	}
	if err != nil {
		return nil, err
	}
	if !arg.IsName() {
		return nil, fmt.Errorf("invalid input: expected command")
	}
	node, ok := s[arg.Raw]
	if !ok {
		return nil, ErrNoCommand
	}

	for len(node.SubCommands) > 0 {
		nx, err := p.Peek()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !nx.IsName() {
			break
		}
		child, ok := node.SubCommands[nx.Raw]
		if !ok {
			break
		}
		p.Commit()
		node = child
	}

	if node.Resolve == nil {
		return nil, ErrNoResolver
	}

	return node.Resolve(p)
}

type peekIter struct {
	it   Iterator
	have bool
	a    Argument
	err  error
}

func (p *peekIter) Peek() (Argument, error) {
	if p.have {
		return p.a, p.err
	}
	p.a, p.err = p.it.Next()
	p.have = true
	return p.a, p.err
}

func (p *peekIter) Commit() {
	p.have = false
}

func (p *peekIter) Next() (Argument, error) {
	if p.have {
		p.have = false
		return p.a, p.err
	}
	return p.it.Next()
}
