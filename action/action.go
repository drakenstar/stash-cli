package action

import (
	"errors"
	"fmt"
)

var ErrEmpty = errors.New("Empty command")
var ErrIncompleteArgument = errors.New("incomplete argument")

type Action struct {
	Name      string
	Arguments ArgumentList
}

func Parse(input string) (*Action, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, err
	}
	p := &parseState{
		in: input,
		t:  tokens,
	}
	err = p.parseCommand()
	if err != nil {
		return nil, err
	}
	return &p.a, nil
}

type parseState struct {
	in string
	t  []Token
	a  Action
	i  int
}

func (p *parseState) remaining() bool {
	return p.i < len(p.t)
}

func (p *parseState) current() *Token {
	if !p.remaining() {
		return nil
	}
	return &p.t[p.i]
}

func (p *parseState) next() *Token {
	p.i++
	return p.current()
}

func (p *parseState) consumeSpace() {
	token := p.next()
	for token != nil && token.Type == TOKEN_WHITESPACE {
		token = p.next()
	}
}

func (p *parseState) parseCommand() error {
	token := p.current()
	if token == nil {
		return ErrEmpty
	}

	if token.Type != TOKEN_IDENTIFIER {
		return errorSyntax(TOKEN_IDENTIFIER, *token)
	}
	p.a.Name = token.String()

	p.consumeSpace()

	for p.remaining() {
		err := p.parseArgument()
		if err != nil {
			return err
		}
		p.consumeSpace()
	}

	return nil
}

func (p *parseState) parseArgument() error {
	start := p.current()
	if start.Type != TOKEN_IDENTIFIER && start.Type != TOKEN_STRING && start.Type != TOKEN_QUOTED_STRING {
		return errorSyntax(TOKEN_IDENTIFIER, *start)
	}

	next := p.next()
	if next == nil || next.Type != TOKEN_ARGUMENT_SEPARATOR {
		p.a.Arguments = append(p.a.Arguments, ArgumentValue{
			Name:  "",
			Value: start.String(),
		})
		return nil
	}

	if start.Type != TOKEN_IDENTIFIER {
		return errorSyntax(TOKEN_IDENTIFIER, *start)
	}

	value := p.next()
	if value == nil {
		return ErrIncompleteArgument
	}
	if value.Type != TOKEN_IDENTIFIER && value.Type != TOKEN_STRING && value.Type != TOKEN_QUOTED_STRING {
		return errorSyntax(TOKEN_IDENTIFIER, *value)
	}
	p.a.Arguments = append(p.a.Arguments, ArgumentValue{
		Name:  start.String(),
		Value: value.String(),
	})

	return nil
}

func errorSyntax(expected TokenType, actual Token) error {
	return fmt.Errorf("invalid token found as pos %d: %s(%s), expected: %s", actual.Pos, actual.Type, actual, expected)
}
