package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"
	"bufio"

	"github.com/chromedp/chromedp"
)

func timedRun(ctx context.Context, task chromedp.Action) (error, float64) {
	start := time.Now()
	err := chromedp.Run(ctx, task)
	return err, time.Since(start).Seconds()
}

func main() {
	cliInput := flag.String("in", "test.log", "Browser test playbook")
	cliDebug := flag.Bool("debug", false, "enable debugging")
	flag.Parse()

	opts := chromedp.DefaultExecAllocatorOptions[:]

	if *cliDebug {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	bctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(bctx)
	defer cancel()

	f, err := os.Open(*cliInput)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		task, err := parse(scanner.Text()) // Println will add back the final '\n'
		if err != nil {
			log.Fatal(err)
		}
		err, secs := timedRun(ctx, task)
		if err != nil {
			if *cliDebug {
				time.Sleep(time.Second * 10)
			}
			log.Fatal(err)
		}
		log.Println(task, secs)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
