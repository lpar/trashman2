package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/xattr"
)

const xattrKey = "user.com.ath0.tm.date"

const whitespace = "\t\n\x00 "

var verbose = false
var dryrun = false

func getcrtime(fname string) (time.Time, error) {
	xdt, err := xattr.Getxattr(fname, xattrKey)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no data available") {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	dts := strings.Trim(string(xdt), whitespace)
	dt, err := time.Parse(time.RFC3339, dts)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "can't parse %s: %s", dts, err)
		return time.Time{}, err
	}
	// fmt.Printf("parsed %s as %s", dts, dt.String())
	return dt, nil
}

func setcrtime(fname string, timestamp time.Time) error {
	dts := timestamp.Format(time.RFC3339)
	err := xattr.Setxattr(fname, xattrKey, []byte(dts))
	return err
}

func getsetcrtime(fname string, timestamp time.Time) (time.Time, error) {
	t, err := getcrtime(fname)
	if t.IsZero() {
		t = time.Now()
		err2 := setcrtime(fname, timestamp)
		if err2 != nil {
			return t, err2
		}
	}
	return t, err
}

func process(fname string, dur time.Duration, override time.Time, depth int) error {
	f, err := os.Stat(fname)
	if err != nil {
		return err
	}
	if f.Mode().IsDir() {
		return processDir(fname, dur, override, depth)
	}
	return processFile(fname, dur, false)
}

func ageString(d time.Duration) string {
	return fmt.Sprintf("%d days", d/time.Hour/24)
}

func processFile(fname string, dur time.Duration, isdir bool) error {
	t, err := getsetcrtime(fname, time.Now())
	if err != nil {
		return err
	}
	age := time.Since(t)
	if age > dur {
		if verbose {
			fmt.Printf("  rm %s: %s old\n", fname, ageString(age))
		}
		if !dryrun {
			var err2 error
			if isdir {
				err2 = os.RemoveAll(fname)
			} else {
				err2 = os.Remove(fname)
			}
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "error: %s", err2)
			}
		}
	} else {
		if verbose {
			fmt.Printf("keep %s: %s old\n", fname, ageString(age))
		}
	}
	return nil
}

func processDir(fname string, dur time.Duration, override time.Time, depth int) error {
	if depth > 0 {
		// Treat like a file
		err := processFile(fname, dur, true)
		return err
	}
	// Else recurse
	dir, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer dir.Close()
	files, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	var cerr error = nil
	for _, file := range files {
		path := filepath.Join(fname, file)
		err2 := process(path, dur, override, depth+1)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "error processing directory: %s\n", err2)
			cerr = err2
		}
	}
	return cerr
}

func main() {

	pdur := flag.String("days", "14", "days to keep files for before deleting")
	flag.BoolVar(&verbose, "verbose", false, "display verbose messages")
	flag.BoolVar(&dryrun, "dryrun", false, "don't actually delete anything")
	flag.Parse()

	d, err := strconv.Atoi(*pdur)
	if err != nil {
		log.Fatal(err)
	}
	dur := time.Duration(d) * 24 * time.Hour

	for _, fname := range flag.Args() {
		err := process(fname, dur, time.Time{}, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
		}
	}

}
