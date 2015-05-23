package moon

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func runEvalTest(t *testing.T, basepath, inpath, outpath string) {
	in, err := os.Open(inpath)
	if err != nil {
		t.Errorf("unable to open input file %s: %s", inpath, err)
		return
	}
	defer in.Close()

	out, err := os.Open(outpath)
	if err != nil {
		t.Errorf("unable to open output file %s: %s", outpath, err)
		return
	}
	defer out.Close()

	r_inpath := filepath.Base(inpath)
	n, err := strconv.ParseInt(strings.TrimSuffix(r_inpath, ".in"), 10, 64)
	if err != nil {
		t.Errorf("unable to get test number for path %s: %s", inpath, err)
		return
	}

	inDoc, err := Read(in)
	if err != nil {
		t.Errorf("unable to read moon doc from infile: %s", err)
		return
	}

	outDoc, err := Read(out)
	if err != nil {
		t.Errorf("unable to read moon doc from outfile: %s", err)
		return
	}
	if !reflect.DeepEqual(inDoc, outDoc) {
		t.Errorf("test %d: input and output documents do not match!", n)
		t.Logf("input document: %v", inDoc)
		t.Logf("output document: %v", outDoc)
	}
}

func TestEval(t *testing.T) {
	files, err := filepath.Glob("tests/eval/*.in")
	if err != nil {
		t.Errorf("unable to find test files: %s", err)
		return
	}

	for _, fname := range files {
		runEvalTest(t, "tests/eval/", fname, strings.Replace(fname, "in", "out", -1))
	}
}
