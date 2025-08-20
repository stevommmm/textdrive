package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

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

	buf := ""
	fscan := bufio.NewScanner(f)
	for fscan.Scan() {
		buf += fscan.Text()
		task, err := parse(buf) // Println will add back the final '\n'
		if err != nil {
			if errors.Is(err, ErrReadMore) {
				continue
			}
			log.Fatalf("Unrecoverable parser error %s\n", err)
		}
		buf = ""
		err, secs := timedRun(ctx, task)
		if err != nil {
			chromedp.Run(ctx, &Screenshot{Name: "fatal.png"})
			log.Fatal(err)
		}
		log.Println(task, secs)
	}
	if err := fscan.Err(); err != nil {
		log.Fatal(err)
	}
}
