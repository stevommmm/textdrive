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
	ClassName string  `xml:"classname"`
	Name      string  `xml:"name"`
	Time      float64 `xml:"time"`
}

type TestSuites struct {
	Disabled int     `xml:"disabled"`
	Failures int     `xml:"failures"`
	Tests    int     `xml:"tests"`
	Time     float64 `xml:"time"`
	Name     string  `xml:"name"`
	Cases    []TestCase
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

	tc := &TestSuites{}
	xo, err := os.Create("report.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer xo.Close()
	reportencoder := xml.NewEncoder(xo)

	execerr := func() error {
		defer reportencoder.Encode(tc)
		// Read each line into a buffer to process, actions can span multiple lines
		// in which case parse() will tell us to slurp more data via error
		buf := ""
		fscan := bufio.NewScanner(f)
		for fscan.Scan() {
			buf += fscan.Text()
			task, err := parse(buf) // try parse the provided text as go
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
			tc.Tests += 1
			tc.Time += secs
			tc.Cases = append(tc.Cases, TestCase{ClassName: fmt.Sprintf("%s", task), Time: secs})
			if err != nil {
				tc.Failures += 1
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
		log.Fatal(err)
	}
}
