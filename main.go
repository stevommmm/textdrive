package main

import (
	"context"
	"flag"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-logfmt/logfmt"
)

func resolveAction(name string) chromedp.Action {
	switch name {
	case "load":
		return &RemoteLoad{Timeout: "60s"}
	case "click":
		return &RemoteClick{Timeout: "60s"}
	case "type":
		return &RemoteType{Timeout: "60s"}
	case "submit":
		return &RemoteSubmit{Timeout: "60s"}
	case "value":
		return &RemoteValue{Timeout: "60s"}
	case "wait":
		return &RemoteWait{Timeout: "60s"}
	case "scroll":
		return &RemoteScroll{Timeout: "60s"}
	case "compare":
		return &RemoteCompare{Timeout: "60s"}
	case "sleep":
		return &Sleep{Timeout: "60s"}
	default:
		return &Noop{}
	}
}

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

	scanner := logfmt.NewDecoder(f)
	for scanner.ScanRecord() {
		scanner.ScanKeyval()
		taskname := string(scanner.Key())
		if string(scanner.Value()) != "" {
			log.Fatalf("Invalid arguement to task name %q", taskname)
		}
		task := resolveAction(taskname)
		if task == nil {
			log.Fatalf("Unknown task name %q", taskname)
		}
		for scanner.ScanKeyval() {
			k := string(scanner.Key())
			v := string(scanner.Value())

			s := reflect.ValueOf(task).Elem()
			f := s.FieldByName(k)
			if !f.IsValid() {
				log.Fatalf("Invalid field given for %q %q=%q", task, k, v)
			}
			f.SetString(v)
		}
		err, secs := timedRun(ctx, task)
		if err != nil {
			if *cliDebug {
				time.Sleep(time.Second*10)
			}
			log.Fatal(err)
		}
		log.Println(task, secs)
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
}
