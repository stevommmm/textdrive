package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func mustParseDuration(dur string) time.Duration {
	d, err := time.ParseDuration(dur)
	if err != nil {
		panic(err)
	}
	return d
}

// Specific implementation of actions
type RemoteLoad struct {
	Timeout string
	Url     string
}

func (s RemoteLoad) String() string {
	return fmt.Sprintf("load:on:%q", s.Url)
}

func (s RemoteLoad) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.Navigate(s.Url).Do(actionctx)
}

type RemoteClick struct {
	Timeout  string
	Selector string
}

func (s RemoteClick) String() string {
	return fmt.Sprintf("click:on:%q", s.Selector)
}

func (s RemoteClick) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.Click(s.Selector, chromedp.ByQuery).Do(actionctx)
}

type RemoteType struct {
	Timeout  string
	Selector string
	Text     string
}

func (s RemoteType) String() string {
	return fmt.Sprintf("type:%q:in:%q", s.Text, s.Selector)
}

func (s RemoteType) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.SendKeys(s.Selector, s.Text, chromedp.ByQuery).Do(actionctx)
}

type RemoteSubmit struct {
	Timeout  string
	Selector string
}

func (s RemoteSubmit) String() string {
	return fmt.Sprintf("submit:on:%q", s.Selector)
}

func (s RemoteSubmit) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.Submit(s.Selector, chromedp.ByQuery).Do(actionctx)
}

type RemoteWait struct {
	Timeout  string
	Selector string
}

func (s RemoteWait) String() string {
	return fmt.Sprintf("wait:on:%q", s.Selector)
}

func (s RemoteWait) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.WaitVisible(s.Selector, chromedp.ByQuery).Do(actionctx)
}

type RemoteScroll struct {
	Timeout  string
	Selector string
}

func (s RemoteScroll) String() string {
	return fmt.Sprintf("scroll:to:%q", s.Selector)
}

func (s RemoteScroll) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.ScrollIntoView(s.Selector, chromedp.ByQuery).Do(actionctx)
}

type RemoteValue struct {
	Timeout  string
	Selector string
	Value    string
}

func (s RemoteValue) String() string {
	return fmt.Sprintf("value:%q:in:%q", s.Value, s.Selector)
}

func (s RemoteValue) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()
	return chromedp.SetValue(s.Selector, s.Value, chromedp.ByQuery).Do(actionctx)
}

//

type Noop struct{}

func (s Noop) String() string {
	return "noop"
}

func (s Noop) Do(ctx context.Context) error {
	return nil
}

type Sleep struct {
	Timeout string
}

func (s Sleep) String() string {
	return fmt.Sprintf("sleep:for:%s", s.Timeout)
}

func (s Sleep) Do(ctx context.Context) error {
	time.Sleep(mustParseDuration(s.Timeout))
	return nil
}

type Debug struct{}

func (s Debug) String() string {
	return "debug"
}

func (s Debug) Do(ctx context.Context) error {
	var buf []byte
	chromedp.CaptureScreenshot(&buf).Do(ctx)
	os.WriteFile("debug.png", buf, 0o644)
	return nil
}

// Comparison functions

type RemoteCompare struct {
	Timeout  string
	Selector string
	Value    string
}

func (s RemoteCompare) String() string {
	return fmt.Sprintf("compare:%q:to:%q", s.Value, s.Selector)
}

func (s RemoteCompare) Do(ctx context.Context) error {
	actionctx, cancel := context.WithTimeout(ctx, mustParseDuration(s.Timeout))
	defer cancel()

	var val string
	if err := chromedp.TextContent(s.Selector, &val, chromedp.ByQuery).Do(actionctx); err != nil {
		return err
	}
	if strings.EqualFold(val, s.Value) {
		return nil
	}
	return fmt.Errorf("Selector %q value %q does not match %q", s.Selector, val, s.Value)
}
