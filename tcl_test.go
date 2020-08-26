// Copyright 2020 The Sqlite Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite // import "modernc.org/sqlite"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"modernc.org/tcl"
)

var (
	oMaxError = flag.Uint("maxerror", 0, "argument of -maxerror passed to the Tcl test suite")
	oStart    = flag.String("start", "", "argument of -start passed to the Tcl test suite (--start=[$permutation:]$testfile)")
	oTclSuite = flag.String("suite", "full", "Tcl test suite [test-file] to run")
	oVerbose  = flag.String("verbose", "0", "argument of -verbose passed to the Tcl test suite, must be set to a boolean (0, 1) or to \"file\"")
)

func TestTclTest(t *testing.T) {
	blacklist := map[string]struct{}{}
	m, err := filepath.Glob(filepath.FromSlash("testdata/tcl/*"))
	if err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "sqlite-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)

	testfixture := filepath.Join(dir, "testfixture")
	args0 := []string{"build", "-o", testfixture}
	tags := "-tags=libc.nofsync"
	if s := *oXTags; s != "" {
		tags += "," + s
	}
	args0 = append(args0, tags, "modernc.org/sqlite/internal/testfixture")
	cmd := exec.Command("go", args0...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s\n%v", out, err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	for _, v := range m {
		if _, ok := blacklist[filepath.Base(v)]; ok {
			continue
		}

		s := filepath.Join(wd, v)
		d := filepath.Join(dir, filepath.Base(v))
		f, err := ioutil.ReadFile(s)
		if err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(d, f, 0660); err != nil {
			t.Fatal(err)
		}
	}

	library := filepath.Join(dir, "library")
	if err := tcl.Library(library); err != nil {
		t.Fatal(err)
	}

	os.Setenv("TCL_LIBRARY", library)
	var args []string
	switch s := *oTclSuite; s {
	case "":
		args = []string{"all.test"}
	default:
		a := strings.Split(s, " ")
		args = append([]string{"permutations.test"}, a...)
	}
	if *oVerbose != "" {
		args = append(args, fmt.Sprintf("-verbose=%s", *oVerbose))
	}
	if *oMaxError != 0 {
		args = append(args, fmt.Sprintf("-maxerror=%d", *oMaxError))
	}
	if *oStart != "" {
		args = append(args, fmt.Sprintf("-start=%s", *oStart))
	}
	switch runtime.GOOS {
	case "windows":
		panic(todo(""))
	default:
		os.Setenv("PATH", fmt.Sprintf("%s:%s", dir, os.Getenv("PATH")))
	}
	cmd = exec.Command(testfixture, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}
