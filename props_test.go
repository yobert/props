package props

import (
	"bytes"
	"testing"
)

func TestProps(t *testing.T) {

	filename := "./testdata/1.properties"
	data, err := DecodeFile(filename)
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

	for k, v := range data {
		w, ok := want[k]
		if !ok {
			t.Fatalf("for key %#v got unexpected value %#v at %s:%d", k, v.Value, filename, v.SourceLine)
		}
		if w != v.Value {
			t.Fatalf("for key %#v got %#v but expected %#v at %s:%d", k, v.Value, w, filename, v.SourceLine)
		}
		delete(want, k)
	}
	for k, v := range want {
		t.Fatalf("never found wanted key %#v value %#v in %s", k, v.Value, filename)
	}
}

func TestWrite(t *testing.T) {
	buf := &bytes.Buffer{}

	e := NewEncoder(buf)

	err := e.Encode(&Chunk{Key: "test", Value: "stuff"})
	if err != nil {
		t.Fatal(err)
	}
	err = e.Encode(&Chunk{Key: "", Value: "# here be a comment\n"})
	if err != nil {
		t.Fatal(err)
	}
	err = e.Encode(&Chunk{Key: "escaped\r\n", Value: "things\t\t\\"})
	if err != nil {
		t.Fatal(err)
	}

	want := `test=stuff
# here be a comment
escaped\r\n=things\t\t\\
`

	got := buf.String()

	if got != want {
		t.Log("got:\n" + got)
		t.Log("wanted:\n" + want)
		t.Fatal("encoding failed")
	}
}
