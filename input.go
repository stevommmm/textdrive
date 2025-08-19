package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"reflect"
	"strconv"

	"github.com/chromedp/chromedp"
)

func parse(expr string) (chromedp.Action, error) {
	e, err := parser.ParseExpr(expr)
	if err != nil {
		log.Printf("bad: %#v\n", e)
		return nil, err
	}
	log.Printf("ok: %#v\n", e)
	x, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("Invalid definition found, expected thing{args:value}")
	}
	act := resolveAction(x.Type.(*ast.Ident).Name)
	actref := reflect.ValueOf(act).Elem()
	for _, v := range x.Elts {
		f, ok := v.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("No arguements given to type %q", x.Type)
		}
		k := f.Key.(*ast.Ident)
		v := f.Value.(*ast.BasicLit)
		if v.Kind != token.STRING {
			return nil, fmt.Errorf("Attribute %q on type %q is not a string", k, x.Type)
		}
		s, err := strconv.Unquote(v.Value)
		if err != nil {
			return nil, err
		}
		fnc := actref.FieldByName(k.Name)
		if !fnc.IsValid() {
			return nil, fmt.Errorf("Invalid field given for %q %q=%q", act, k, v)
		}
		fnc.SetString(s)
	}
	return act, nil
}
