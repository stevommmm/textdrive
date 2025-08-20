# TextDrive

Browser automation via simple text actions.

Actions are defined as Golang struct definitions, with case insensitive field names and no overarching package information.

The following actions and attributes are provided:

- `load`:
	- `url`: Destination to navigate to
	- `timeout`: A valid `time.Duration` string
- `click`:
	- `selector`: A CSS selector to mouse click on
	- `timeout`: A valid `time.Duration` string
- `type`:
	- `selector`: A CSS selector to input text to
	- `text`: Character value to input via keys
	- `timeout`: A valid `time.Duration` string
- `submit`:
	- `selector`: A CSS selector to trigger submission on
	- `timeout`: A valid `time.Duration` string
- `value`:
	- `selector`: A CSS selector to mouse click on
	- `value`: Character value to set
	- `timeout`: A valid `time.Duration` string
- `wait`:
	- `selector`: An **optional** CSS selector to wait for load of
	- `timeout`: A valid `time.Duration` string
- `scroll`:
	- `selector`: A CSS selector to scroll into focus
	- `timeout`: A valid `time.Duration` string
- `compare`:
	- `selector`: A CSS selector to resolve text content fo
	- `value`: Expected character value to match
	- `timeout`: A valid `time.Duration` string
- `record`:
	- `selector`: A CSS selector to log the text content of
	- `timeout`: A valid `time.Duration` string
- `screenshot`:
	- `name`: File path to save the recorded full page screenshot


### Example
```go
load{url: "https://iana.org/"}
wait{
	Selector: "body",
}
click{Selector: "a[href='/domains']"}
wait{Timeout: "4s"} // Wait can be a time or selector
click{Selector: "p > a[href='/domains/special']"}
wait{Selector: "article > main"}
compare{
	Selector: "article > main h2:last-of-type",
	Value: "Other Special-Use Domains",
}
screenshot{name: "capture.png"}
```

```text
2025/08/20 10:20:16 load:on:"https://iana.org/" 5.142491329
2025/08/20 10:20:16 wait:on:"body" 0.00556122
2025/08/20 10:20:16 click:on:"a[href='/domains']" 0.010442514
2025/08/20 10:20:20 wait:for:"4s" 4.001158436
2025/08/20 10:20:20 click:on:"p > a[href='/domains/special']" 0.011963214
2025/08/20 10:20:20 wait:on:"article > main" 0.854157673
2025/08/20 10:20:21 compare:"Other Special-Use Domains":to:"article > main h2:last-of-type" 0.006131768
2025/08/20 10:20:22 screenshot:"capture.png" 1.407466757
```
