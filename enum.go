// Package enums helps you find instances when you find cases where you
// have not updated code to handle all enums you have. This is not a fully
// automated like exhaustive(https://github.com/nishanths/exhaustive) but
// rather intended to be used in cases where you might want to apply your own
// logic while still being warned when new cases pop up.
//
// Enums supports basic literals and structs tagged with `enum:"identifier"`.
//
// The canonical example of ths package is that you maintain feature flags
// in your codebase, and you have a function that returns all flags that are
// relevant for a particular section of code. As any new flag is introduced
// they need to be considered for inclusion, and it's nice if the test suite
// enforces the consideration.
//
// The entry point for enums is the All function which takes a package path
// and a type and will return a Collection of found instances of that type.
// A collection can produce a Diff given a slice of objects.
//
// There is a helper for the standard case of "match all of this type" in
// enumstest.NoDiff.
package enums

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Collection contains found matches from All and can be diffed against values.
type Collection struct {
	Type      string // the import path of the type
	FieldName string // if the underlying type is a struct this value is the name of the field that is used to distinguish flags
	Enums     []Enum // all distinct values found
}

// Enum represents a value for a matched type.
//
// Example:
//   var MyFlag Flag = "Hello"
// Is equivalent to:
//   Enum{Name: "MyFlag", Value: "Hello"}
type Enum struct {
	Name  string
	Value string
}

// All finds variables of typ in pkg.
//
// Example:
//   All("./feature", "feature.Flag")
func All(pkg string, typ string) (Collection, error) {
	cfg := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName}
	pkgs, err := packages.Load(&cfg, pkg)
	if err != nil {
		return Collection{}, fmt.Errorf("failed to load package: %w", err)
	}

	var collection Collection
	for _, p := range pkgs {
		for e, t := range p.TypesInfo.Defs {
			if t == nil {
				continue
			}

			if !strings.HasSuffix(t.Type().String(), typ) {
				continue
			}

			// A declaration with a value and not a type declaration
			decl, ok := e.Obj.Decl.(*ast.ValueSpec)
			if !ok {
				continue
			}

			var fieldName string
			var val string
			for _, v := range decl.Values {
				switch value := v.(type) {
				case *ast.BasicLit:
					val = value.Value
				case *ast.CompositeLit:
					fieldName, val, err = structValue(value)
					if err != nil {
						return Collection{}, err
					}
				default:
					// Either a case where it would be hard to distinguish or something not considered so far. Likely the latter.
					panic(fmt.Sprintf("unknown type, please file a bug report with example code: '%T'", v))
				}
			}

			collection.Type = t.Type().String()
			collection.FieldName = fieldName
			collection.Enums = append(collection.Enums, Enum{
				Name:  t.Name(),
				Value: val,
			})
		}
	}

	// The values comes out in different order and it made some tests flaky
	sort.Slice(collection.Enums, func(i, j int) bool { return collection.Enums[i].Name < collection.Enums[j].Name })

	return collection, nil
}

func structValue(exp *ast.CompositeLit) (fieldName string, val string, err error) {
	typ := exp.Type.(*ast.Ident)
	decl := typ.Obj.Decl.(*ast.TypeSpec)
	struc := decl.Type.(*ast.StructType)

	for i, f := range struc.Fields.List {
		if strings.Contains(f.Tag.Value, "`enums:\"identifier\"`") {
			if len(f.Names) > 1 {
				// No idea if or how this could happen, so let's ask for help
				panic(fmt.Errorf("struct identifier field has more than one Names, please file a bug report with example code: %#v", f.Names))
			}
			fieldName = f.Names[0].String()

			kv := exp.Elts[i].(*ast.KeyValueExpr)
			fieldVal, ok := kv.Value.(*ast.BasicLit)
			if !ok {
				return "", "", fmt.Errorf("struct identifier value not a basic literal: %s = %#v", fieldName, kv.Value)
			}

			val = fieldVal.Value
			break
		}
	}

	if val == "" {
		return "", "", errors.New(`no struct tag with enum:"identifier" found`)
	}

	return fieldName, val, err
}

// Diff contains the result of checking the difference between a Collection and a list of values.
type Diff struct {
	Missing Collection
	Extra   []string
}

// Zero returns whether there is nothing in the diff.
func (d Diff) Zero() bool {
	return len(d.Missing.Enums) == 0 && len(d.Extra) == 0
}

// String outputs a human summary of the values in the diff.
func (d Diff) String() string {
	var msg string

	if len(d.Missing.Enums) > 0 {
		msg += "Enums declared but not part of actual:\n"
		for _, v := range d.Missing.Enums {
			msg += fmt.Sprintf("\t%s = %s\n", v.Name, v.Value)
		}
	}

	if len(d.Extra) > 0 {
		msg += "Extra values provided but not part of Enums:\n"
		for _, v := range d.Extra {
			msg += fmt.Sprintf("\t%s\n", v)
		}
	}

	if len(msg) > 0 {
		return msg
	}

	return "<Diff{}>"
}

// Diff indicates differences between a collection and any slice.
//
// Because a Collection stores all values as strings the difference is
// calculated based on the string representation of the value.
func (c Collection) Diff(actual interface{}) Diff {
	acTyp := reflect.ValueOf(actual)
	if acTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Diff: actual is not a slice: %T", actual))
	}

	values := make(map[string]Enum, len(c.Enums))
	for _, v := range c.Enums {
		values[v.Value] = v
	}

	var diff Diff
	for i := 0; i < acTyp.Len(); i++ {
		item := acTyp.Index(i)
		val := c.valueFrom(item)

		if _, ok := values[val]; ok {
			delete(values, val)
			continue
		}

		diff.Extra = append(diff.Extra, val)
	}

	diff.Missing = Collection{
		Type:      c.Type,
		FieldName: c.FieldName,
	}
	for _, v := range values {
		diff.Missing.Enums = append(diff.Missing.Enums, v)
	}

	return diff
}

func (c Collection) valueFrom(item reflect.Value) string {
	var val string

	switch item.Type().Kind() {
	case reflect.Struct:
		val = c.fieldValue(item)

		if val == "" {
			val = fmt.Sprintf("%#v", item.Interface())
		}
	default:
		val = fmt.Sprintf("%#v", item.Interface())
	}

	return val
}

func (c Collection) fieldValue(item reflect.Value) string {
	if len(c.Enums) == 0 {
		panic("Diff: collection is empty")
	}

	typ := item.Type()

	if !strings.HasSuffix(c.Type, typ.String()) {
		return ""
	}

	field, ok := typ.FieldByName(c.FieldName)
	if !ok {
		return ""
	}

	if tag, ok := field.Tag.Lookup("enums"); !ok || tag != "identifier" {
		return ""
	}

	return fmt.Sprintf("%#v", item.FieldByName(c.FieldName).Interface())
}
