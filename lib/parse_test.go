package moon

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runParseTest(t *testing.T, basepath, inpath, outpath string) {
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
	n, err := strconv.ParseInt(strings.TrimSuffix(r_inpath, ".in"), 0, 64)
	if err != nil {
		t.Errorf("unable to get test number for path %s: %s", inpath, err)
		return
	}

	var buf bytes.Buffer
	root, err := parse(in)
	if err != nil {
		t.Logf("test %d: in: %s out: %s", n, inpath, outpath)
		t.Errorf("parse error in test %d: %s", n, err)
		return
	}
	if err := root.pretty(&buf, ""); err != nil {
		t.Logf("test %d: in: %s out: %s", n, inpath, outpath)
		t.Errorf("output error in test %d: %s", n, err)
		return
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Logf("test %d: in: %s out: %s", n, inpath, outpath)
		t.Errorf("lex output does not match expected result for test %d", n)
		t.Logf("expected output:\n%s", expected)
		t.Logf("received output:\n%s", buf.Bytes())
	}
}

func TestParse(t *testing.T) {
	files, err := filepath.Glob("tests/parse/*.in")
	if err != nil {
		t.Errorf("unable to find test files: %s", err)
		return
	}

	for _, fname := range files {
		runParseTest(t, "tests/parse/", fname, strings.Replace(fname, "in", "out", -1))
	}
}
