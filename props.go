package props

import (
	"fmt"
	"github.com/bjarneh/latinx"
	"io"
	"os"
)

type Props struct {
	File string
	Data map[string]string
	Line map[string]int

	state  pState
	escape bool
	lin    int
	key    []byte
	val    []byte
}

type pState int

const (
	noneState pState = iota
	commentState
	keyState
	keyPostState // key terminated, haven't found : or = yet
	valLeadState // in whitespace we shouldn't include in the value
	valState
)

func New() *Props {
	return &Props{
		Data: make(map[string]string),
		Line: make(map[string]int),
	}
}

func (p *Props) String(key string) string {
	return p.Data[key]
}
func (p *Props) Source(key string) string {
	line, ok := p.Line[key]
	if ok {
		return fmt.Sprintf("%s:%d", p.File, line)
	}
	return p.File
}

func (p *Props) ParseFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	p.File = file

	p.state = noneState
	p.lin = 0

	// java standard is ISO 8859-1 for properties files.
	// it's dumb but whatever.

	r := latinx.NewReader(latinx.ISO_8859_1, f)

	return p.parse(r)
}

func (p *Props) parse(r io.Reader) error { // parse assumes utf8 coming in
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				p.consume(b)
			}
		}
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (p *Props) consume(b byte) {
	if b == '\n' {
		p.lin++
	} else if b == '\r' {
		return
	}

	switch p.state {
	case noneState:
		if p.escape {
			p.escape = false
			if b == '\n' {
				// do nothing, we haven't started a key yet
			} else {
				p.key = nil
				p.val = nil
				b = unescape(b)
				p.key = append(p.key, b)
				p.state = keyState
			}
		} else {
			switch b {
			case ' ', '\t', '\f', '\n':
			case '!', '#':
				p.state = commentState
			case '\\':
				p.escape = true
			default:
				p.key = nil
				p.val = nil
				p.key = append(p.key, b)
				p.state = keyState
			}
		}

	case commentState:
		switch b {
		case '\n':
			p.state = noneState
		}

	case keyState:
		if p.escape {
			p.escape = false
			if b == '\n' {
				// keep reading as key
			} else {
				b = unescape(b)
				p.key = append(p.key, b)
			}
		} else {
			switch b {
			case ' ', '\t', '\f':
				p.state = keyPostState
			case ':', '=':
				p.state = valLeadState
			case '\n':
				p.store()
				p.state = noneState
			case '\\':
				p.escape = true
			default:
				p.key = append(p.key, b)
			}
		}

	case keyPostState:
		if p.escape {
			p.escape = false
			if b == '\n' {
				p.state = valLeadState
			} else {
				b = unescape(b)
				p.val = append(p.val, b)
				p.state = valState
			}
		} else {
			switch b {
			case ' ', '\t', '\f':
			case ':', '=':
				p.state = valLeadState
			case '\n':
				p.store()
				p.state = noneState
			case '\\':
				p.escape = true
			default:
				p.val = append(p.val, b)
				p.state = valState
			}
		}

	case valLeadState:
		if p.escape {
			p.escape = false
			if b == '\n' {
				p.state = valLeadState
			} else {
				b = unescape(b)
				p.val = append(p.val, b)
				p.state = valState
			}
		} else {
			switch b {
			case ' ', '\t', '\f':
			case '\n':
				p.store()
				p.state = noneState
			case '\\':
				p.escape = true
			default:
				p.val = append(p.val, b)
				p.state = valState
			}
		}

	case valState:
		if p.escape {
			p.escape = false
			if b == '\n' {
				p.state = valLeadState
			} else {
				b = unescape(b)
				p.val = append(p.val, b)
			}
		} else {
			switch b {
			case '\n':
				p.store()
				p.state = noneState
			case '\\':
				p.escape = true
			default:
				p.val = append(p.val, b)
			}
		}
	}
}

func unescape(b byte) byte {
	switch b {
	case 'r':
		return '\r'
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'f':
		return '\f'
	default:
		return b
	}
}

func (p *Props) store() {
	k := string(p.key)
	if k != "" {
		p.Data[k] = string(p.val)
		p.Line[k] = p.lin
	}
}
