package main

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	ClassName string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Time      float64  `xml:"time,attr"`
}

type TestSuite struct {
	XMLName  xml.Name   `xml:"testsuite"`
	Disabled int        `xml:"disabled,attr"`
	Failures int        `xml:"failures,attr"`
	Tests    int        `xml:"tests,attr"`
	Time     float64    `xml:"time,attr"`
	Name     string     `xml:"name,attr"`
	Cases    []TestCase `xml:"testcase"`
}

type TestSuites struct {
	XMLName  xml.Name    `xml:"testsuites"`
	Disabled int         `xml:"disabled,attr"`
	Failures int         `xml:"failures,attr"`
	Tests    int         `xml:"tests,attr"`
	Time     float64     `xml:"time,attr"`
	Name     string      `xml:"name,attr"`
	Suites   []TestSuite `xml:"testsuite"`
}

func writeSuite(ts *TestSuite) error {
	tc := &TestSuites{
		Name:   ts.Name,
		Suites: []TestSuite{*ts},
	}
	tc.Failures = ts.Failures
	tc.Tests = ts.Tests
	tc.Time = ts.Time
	xo, err := os.Create("report.xml")
	if err != nil {
		return err
	}
	defer xo.Close()
	xo.Write([]byte(xml.Header))
	reportencoder := xml.NewEncoder(xo)
	reportencoder.Encode(tc)
	return nil
}

// Run a chromedp action and report the duration it took
func timedRun(ctx context.Context, task chromedp.Action) (error, float64) {
	start := time.Now()
	err := chromedp.Run(ctx, task)
	return err, time.Since(start).Seconds()
}

func main() {
	cliInput := flag.String("in", "-", "Browser test playbook")
	cliDebug := flag.Bool("debug", false, "enable debugging")
	cliProxy := flag.String("proxy", "", "HTTP Proxy to use")
	flag.Parse()

	opts := chromedp.DefaultExecAllocatorOptions[:]

	if *cliDebug {
		opts = append(opts, chromedp.Flag("headless", false))
	}
	if *cliProxy != "" {
		opts = append(opts, chromedp.ProxyServer(*cliProxy))
	}

	bctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(bctx)
	defer cancel()

	// Figure out the input, if its a dash use stdin otherwise file
	var f *os.File
	if *cliInput == "-" {
		f = os.Stdin
	} else {
		fh, err := os.Open(*cliInput)
		if err != nil {
			log.Fatal(err)
		}
		f = fh
		defer fh.Close()
	}

	tcs := &TestSuite{
		Name:  *cliInput,
		Cases: []TestCase{},
	}
	defer writeSuite(tcs)

	execerr := func() error {
		// Read each line into a buffer to process, actions can span multiple lines
		// in which case parse() will tell us to slurp more data via error
		buf := ""
		fscan := bufio.NewScanner(f)
		for fscan.Scan() {
			buf += fscan.Text()
			task, name, err := parse(buf) // try parse the provided text as go
			if err != nil {
				// Special case handler where input spans multiple lines
				if errors.Is(err, ErrReadMore) {
					continue
				}
				return err
			}
			buf = ""
			// Execute the requested action in the open? browser
			err, secs := timedRun(ctx, task)
			tcs.Tests += 1
			tcs.Time += secs
			tcs.Cases = append(tcs.Cases, TestCase{ClassName: name, Name: fmt.Sprintf("%s", task), Time: secs})
			if err != nil {
				tcs.Failures += 1
				chromedp.Run(ctx, &Screenshot{Name: "fatal.png"})
				return err
			}
			log.Println(task, secs)
		}
		if err := fscan.Err(); err != nil {
			return err
		}
		return nil
	}()

	if execerr != nil {
		log.Fatal(execerr)
	}
}
