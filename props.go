package props

import (
	"errors"
	"github.com/bjarneh/latinx"
	"io"
	"os"
)

var errDecoderClosed = errors.New("Decode() on closed decoder")

type Decoder struct {
	r io.Reader

	filename string
	state    pState
	escape   bool
	lin      int
	key      []byte
	val      []byte
	cc       chan cce
	cq       chan struct{}
}
type cce struct {
	chunk *Chunk
	err   error
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

type Encoder struct {
	w io.Writer
}

// Chunk holds one key/value pair, for both encoding and decoding.
// if Key is the empty string, Value holds either whitespace or a comment.
type Chunk struct {
	Key        string
	Value      string
	SourceLine int
}

type Map map[string]*Chunk

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, state: noneState}
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func DecodeFile(path string) (Map, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	m := make(Map)
	d := NewDecoder(fh)
	d.filename = path
	for {
		chunk, err := d.Decode()
		if chunk == nil {
			break
		}
		if err != nil {
			return nil, err
		}
		if chunk.Key != "" {
			m[chunk.Key] = chunk
		}
	}
	return m, nil
}

func EncodeFile(path string, data Map) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	e := NewEncoder(fh)
	for _, c := range data {
		err := e.Encode(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Decoder) Decode() (*Chunk, error) {
	if p.cc == nil {
		p.cc = make(chan cce)
		p.cq = make(chan struct{})
		go func() {
			// java standard is ISO 8859-1 for properties files.
			// it's dumb but whatever.
			r := latinx.NewReader(latinx.ISO_8859_1, p.r)

			buf := make([]byte, 1024)

			for {
				n, err := r.Read(buf)
				if n > 0 {
					for _, b := range buf[:n] {
						p.consume(b)
					}
				} else if err != nil {
					p.cc <- cce{nil, err}
					return
				}
			}
		}()
	}
	c, ok := <-p.cc
	if !ok {
		return nil, errDecoderClosed
	}
	if c.err == io.EOF {
		return nil, nil
	}
	return c.chunk, c.err
}

func (p *Decoder) consume(b byte) {
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

func (p *Decoder) store() {
	chunk := Chunk{
		Key:        string(p.key),
		Value:      string(p.val),
		SourceLine: p.lin,
	}
	p.cc <- cce{&chunk, nil}
}
