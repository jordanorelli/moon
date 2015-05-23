package moon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runLexTest(t *testing.T, basepath, inpath, outpath string) {
	in, err := os.Open(inpath)
	if err != nil {
		t.Errorf("unable to open input file %s: %s", inpath, err)
		return
	}
	defer in.Close()

	expected, err := ioutil.ReadFile(outpath)
	if err != nil {
		t.Errorf("unable to read expected output for %s: %s", outpath, err)
		return
	}

	r_inpath := filepath.Base(inpath)
	n, err := strconv.ParseInt(strings.TrimSuffix(r_inpath, ".in"), 10, 64)
	if err != nil {
		t.Errorf("unable to get test number for path %s: %s", inpath, err)
		return
	}

	var buf bytes.Buffer
	c := lex(in)
	for t := range c {
		fmt.Fprintln(&buf, t)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Logf("test %d: in: %s out: %s", n, inpath, outpath)
		t.Errorf("lex output does not match expected result for test %d", n)
		t.Logf("expected output:\n%s", expected)
		t.Logf("received output:\n%s", buf.Bytes())
	}
}

func TestLex(t *testing.T) {
	files, err := filepath.Glob("tests/lex/*.in")
	if err != nil {
		t.Errorf("unable to find test files: %s", err)
		return
	}

	for _, fname := range files {
		runLexTest(t, "tests/lex/", fname, strings.Replace(fname, "in", "out", -1))
	}
}
