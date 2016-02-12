package props

import (
	"io"
	//"fmt"
)

var escapes = map[uint8][]byte{
	'\r': []byte("\\r"),
	'\n': []byte("\\n"),
	'\t': []byte("\\t"),
	'\f': []byte("\\f"),
	'\\': []byte("\\\\"),
}

func write_escaped(w io.Writer, s string) error {
again:

	for i := 0; i < len(s); i++ {
		c := s[i]

		//fmt.Printf("%#v (%T) %#v\n", c, c, string(c))

		if c == '\r' || c == '\n' || c == '\t' || c == '\f' || c == '\\' {
			if i > 0 {
				_, err := w.Write([]byte(s[:i]))
				if err != nil {
					return err
				}
				s = s[i:]
				goto again
			}

			_, err := w.Write(escapes[c])
			if err != nil {
				return err
			}
			s = s[1:]
			goto again
		}
	}

	if len(s) > 0 {
		_, err := w.Write([]byte(s))
		return err
	}
	return nil
}

func WriteTo(w io.Writer, k, v string) error {
	err := write_escaped(w, k)
	if err != nil {
		return err
	}
	err = write_escaped(w, "=")
	if err != nil {
		return err
	}
	err = write_escaped(w, v)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}
