// Package typedecl helps you find cases where you have not updated code to
// handle all variables of a specific type.
// This is not a fully automated like [exhaustive], but rather intended
// to be used in cases where you might want to apply your own logic while
// still being warned when new cases pop up.
//
// typedecl supports basic literals and structs tagged with `typedecl:"identifier"`.
//
// The entry point is the [All] function which takes a package path
// and a type expressed as the "pkgname.Type", it will return
// a [Collection] of found instances of that type.
// A collection can produce a [Diff] given a slice of objects.
//
// There is a helper for the standard case of tell me of all missing
// matches of this type in typedecltest.NoDiff.
//
// [exhaustive]: https://github.com/nishanths/exhaustive
package typedecl

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	structTagName       = `typedecl`
	structTagIdentifier = `identifier`
)

// Collection contains found matches from All and can be diffed against values.
type Collection struct {
	Type      string  // the import path of the type
	FieldName string  // if the underlying type is a struct this value is the name of the field that is used to distinguish flags
	Matches   []Match // all distinct values found
}

// Match represents a value for a matched type.
//
// Example:
//
//	var MyFlag Flag = "Hello"
//
// Is equivalent to:
//
//	Match{Name: "MyFlag", Value: "Hello"}
type Match struct {
	Name  string
	Value string
}

// All finds variables of typ in pkg.
//
// Example:
//
//	All("./feature", "feature.Flag")
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

			// Handling [issue #38] from GitHub where declarations inside of functions would be found and used,
			// whereas we only care about package-level declarations.
			// To be honest, this feels _quite_ hacky but the cases I could think of right now seems covered,
			// so let's see what else we discover with more use.🤷
			//
			// [issue #38]: https://github.com/gaqzi/typedecl/issues/38
			if notTopLevelDeclaration := t.Parent() != nil && t.Parent().Parent() != nil && t.Parent().Parent().Pos() != token.NoPos; notTopLevelDeclaration {
				continue
			}

			// Ignore slices as they're a collection of the enum and therefore most likely a variable used to
			// collect the declared types, and not something we want to use directly.
			if _, ok := t.Type().(*types.Slice); ok {
				continue
			}

			// Ignore types that don't have Obj since they definitely can't be turned into a declaration
			// TODO: Obj is deprecated and there's a new way of getting at this information, look into this new way
			if e.Obj == nil {
				continue
			}

			// A declaration with a value and not a type declaration
			decl, ok := e.Obj.Decl.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Ignore values that don't match the string of the type we're matching against.
			if !strings.HasSuffix(t.Type().String(), typ) {
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
			collection.Matches = append(collection.Matches, Match{
				Name:  t.Name(),
				Value: val,
			})
		}
	}

	// The values come out in different order, and it made some tests flaky
	sort.Slice(collection.Matches, func(i, j int) bool { return collection.Matches[i].Name < collection.Matches[j].Name })

	return collection, nil
}

func structValue(exp *ast.CompositeLit) (fieldName string, val string, err error) {
	typ := exp.Type.(*ast.Ident)
	decl := typ.Obj.Decl.(*ast.TypeSpec)
	struc := decl.Type.(*ast.StructType)

	for i, f := range struc.Fields.List {
		if strings.Contains(f.Tag.Value, "`"+structTagName+":\""+structTagIdentifier+"\"`") {
			if len(f.Names) > 1 {
				// No idea if or how this could happen, so let's ask for help
				panic(fmt.Errorf("struct identifier field has more than one Name, please file a bug report with example code: %#v", f.Names))
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
		return "", "", fmt.Errorf(`no struct tag with %s:"%s" found`, structTagName, structTagIdentifier)
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
	return len(d.Missing.Matches) == 0 && len(d.Extra) == 0
}

// String outputs a human summary of the values in the diff.
func (d Diff) String() string {
	var msg string

	if len(d.Missing.Matches) > 0 {
		msg += "Matches declared but not part of actual:\n"
		for _, v := range d.Missing.Matches {
			msg += fmt.Sprintf("\t%s = %s\n", v.Name, v.Value)
		}
	}

	if len(d.Extra) > 0 {
		msg += "Extra values provided but not part of Matches:\n"
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

	values := make(map[string]Match, len(c.Matches))
	for _, v := range c.Matches {
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
		diff.Missing.Matches = append(diff.Missing.Matches, v)
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
	if len(c.Matches) == 0 {
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

	if tag, ok := field.Tag.Lookup(structTagName); !ok || tag != structTagIdentifier {
		return ""
	}

	return fmt.Sprintf("%#v", item.FieldByName(c.FieldName).Interface())
}
