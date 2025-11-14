package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"reflect"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
)

var ErrReadMore = errors.New("Read more input")

// Case insensitive field resolution by name
func iFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

// Parse the provided line(s) as Go code
func parse(expr string) (chromedp.Action, string, error) {
	e, err := parser.ParseExpr(expr)
	if err != nil {
		// If we got an EOF, try getting more input
		errs, ok := err.(scanner.ErrorList)
		if ok && errs.Len() == 1 && strings.HasSuffix(errs[0].Msg, "found 'EOF'") {
			return nil, "", ErrReadMore
		}
		return nil, "", err
	}
	// We're expecting thing{args:blah}
	x, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil, "", fmt.Errorf("Invalid definition found, expected thing{args:value}")
	}
	typename := x.Type.(*ast.Ident).Name
	// Check we have a think - defaults to noop
	act := resolveAction(typename)
	actref := reflect.ValueOf(act).Elem()
	// Walk the key:value pairs provided in the struct definition
	for _, v := range x.Elts {
		f, ok := v.(*ast.KeyValueExpr)
		if !ok {
			return nil, "", fmt.Errorf("No arguements given to type %q", x.Type)
		}
		// Turn out the key/value into usable types
		k := f.Key.(*ast.Ident)
		v := f.Value.(*ast.BasicLit)
		if v.Kind != token.STRING {
			return nil, "", fmt.Errorf("Attribute %q on type %q is not a string", k, x.Type)
		}
		// Strings are always quoted, unquote for the raw value
		s, err := strconv.Unquote(v.Value)
		if err != nil {
			return nil, "", err
		}
		// Set the k:v attributes on our action type
		fnc := iFieldByName(actref, k.Name)
		if !fnc.IsValid() {
			return nil, "", fmt.Errorf("Invalid field given for %q %q=%q", act, k, v)
		}
		fnc.SetString(s)
	}
	return act, typename, nil
}
