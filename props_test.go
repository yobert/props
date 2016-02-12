package props

import (
	"bytes"
	"testing"
)

func TestProps(t *testing.T) {

	p := New()
	err := p.ParseFile("./testdata/1.properties")
	if err != nil {
		t.Error(err)
	}

	want := map[string]string{
		"a":                 "asparagus",
		"b":                 "banana",
		"c":                 "cheese",
		"d":                 "drumstick",
		"e":                 "eel sauce # sounds gross but it's tasty",
		"Truth1":            "Beauty",
		"Truth2":            "Beauty",
		"Truth3":            "Beauty",
		"fruits":            "apple, banana, pear, cantaloupe, watermelon, kiwi, mango",
		"cheeses":           "",
		" whitespacekey\\ ": " some whitespace \r\n",
		"\n\n\n":            "\r\r\r",
	}

	for k, v := range p.Data {
		w, ok := want[k]
		if !ok {
			t.Fatalf("for key %#v got unexpected value %#v at %s", k, v, p.Source(k))
		}
		if w != v {
			t.Fatalf("for key %#v got %#v but expected %#v at %s", k, v, w, p.Source(k))
		}
		delete(want, k)
	}
	for k, v := range want {
		t.Fatal("never found wanted key %#v value %#v in %s", k, v, p.File)
	}
}

func TestWrite(t *testing.T) {
	buf := &bytes.Buffer{}

	err := WriteTo(buf, "test", "stuff")
	if err != nil {
		t.Fatal(err)
	}
	err = WriteTo(buf, "escaped\r\n", "things\t\t\\")
	if err != nil {
		t.Fatal(err)
	}

	want := `test=stuff
escaped\r\n=things\t\t\\
`

	got := buf.String()

	if got != want {
		t.Log("got:\n" + got)
		t.Log("wanted:\n" + want)
		t.Fatal("encoding failed")
	}
}
