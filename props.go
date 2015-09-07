package props

import (
	"io"
	"os"

	"github.com/bjarneh/latinx"
)

type Props struct {
	File string
	Data map[string]string
	Line map[string]int

	state pState
	lin int
	key []byte
	val []byte
}

type pState int
const (
	NONE pState = iota
	COMMENT
	KEY
	KEY_ESCAPE
	PREVALUE
	PREVALUE_WHITE
	VALUE
	VALUE_ESCAPE
	VALUE_ESCAPE_WHITE
)

func New() *Props {
	return &Props{
		Data: make(map[string]string),
		Line: make(map[string]int),
	}
}

func (p *Props) ParseFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	p.File = file

	p.state = NONE
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
	case NONE:
		switch b {
		case ' ', '\t', '\n':
		case '!', '#':
			p.state = COMMENT
		case '\\':
			p.key = nil
			p.val = nil
			p.state = KEY_ESCAPE
		default:
			p.key = nil
			p.val = nil
			p.key = append(p.key, b)
			p.state = KEY
		}
	case COMMENT:
		switch b {
		case '\n':
			p.state = NONE
		}
		return
	case KEY:
		switch b {
		case ' ', '\t':
			p.state = PREVALUE_WHITE
		case ':', '=':
			p.state = PREVALUE
		case '\n':
			p.store()
			p.state = NONE
		case '\\':
			p.state = KEY_ESCAPE
		default:
			p.key = append(p.key, b)
		}
	case KEY_ESCAPE:
		p.key = append(p.key, b)
		p.state = KEY
	case PREVALUE_WHITE:
		switch b {
		case ' ', '\t':
		case ':', '=':
			p.state = PREVALUE
		case '\n':
			p.store()
			p.state = NONE
		case '\\':
			p.val = nil
			p.state = VALUE_ESCAPE
		default:
			p.val = nil
			p.val = append(p.val, b)
			p.state = VALUE
		}
	case PREVALUE:
		switch b {
		case ' ', '\t':
		case '\n':
			p.store()
			p.state = NONE
		case '\\':
			p.val = nil
			p.state = VALUE_ESCAPE
		default:
			p.val = nil
			p.val = append(p.val, b)
			p.state = VALUE
		}
	case VALUE:
		switch b {
		case '\n':
			p.store()
			p.state = NONE
		case '\\':
			p.state = VALUE_ESCAPE
		default:
			p.val = append(p.val, b)
		}
	case VALUE_ESCAPE:
		p.val = append(p.val, b)
		if b == '\n' {
			p.state = VALUE_ESCAPE_WHITE
		} else {
			p.state = VALUE
		}
	case VALUE_ESCAPE_WHITE:
		switch b {
		case ' ', '\t':
		case '\n':
			p.store()
			p.state = NONE
		case '\\':
			p.state = VALUE_ESCAPE
		default:
			p.val = append(p.val, b)
			p.state = VALUE
		}
	}
}

func (p *Props) store() {
	k := string(p.key)
	p.Data[k] = string(p.val)
	p.Line[k] = p.lin
}
