package enums

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Collection contains found matches from All and can be diffed against values
type Collection []Enum

// Enum represents a a match of a type found in the AST for a package
type Enum struct {
	Type      string
	Name      string
	FieldName string // if the underlying type is a struct this value is the name of the field that is used to distinguish flags
	Value     string // all values are represented as strings from the AST
}

// All finds all variables of typ in pkg
//   pkg is a full or relative path.
//   typ is a type in the form: pkgname.Type
// Example: All("./feature", "feature.Flag")
func All(pkg string, typ string) (Collection, error) {
	cfg := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedName | packages.NeedDeps}
	pkgs, err := packages.Load(&cfg, pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	var targets []Enum

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
						return nil, err
					}
				default:
					// Either a case where it would be hard to distinguish or something not considered so far. Likely the latter.
					panic(fmt.Sprintf("unknown type, please file a bug report with example code: '%T'", v))
				}
			}

			targets = append(targets, Enum{
				Type:      t.Type().String(),
				Name:      t.Name(),
				FieldName: fieldName,
				Value:     val,
			})
		}
	}

	return targets, nil
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

// Diff contains the result of checking the difference between a Collection and a list of values
type Diff struct {
	Missing Collection
	Extra   []string
}

// Zero returns whether there is nothing in the diff
func (d Diff) Zero() bool {
	return len(d.Missing) == 0 && len(d.Extra) == 0
}

// String outputs a human summary of the values in the diff
func (d Diff) String() string {
	var msg string

	if len(d.Missing) > 0 {
		msg += "Enums declared but not part of actual:\n"
		for _, v := range d.Missing {
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

// Diff indicates differences between a collection and any slice
// Because a Collection stores all values as strings the difference is
// calculated based on the string representation of the value.
func (e Collection) Diff(actual interface{}) Diff {
	acTyp := reflect.ValueOf(actual)
	if acTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Diff: actual is not a slice: %T", actual))
	}

	values := make(map[string]Enum, len(e))
	for _, v := range e {
		values[v.Value] = v
	}

	var diff Diff
	for i := 0; i < acTyp.Len(); i++ {
		item := acTyp.Index(i)
		val := e.valueFrom(item)

		if _, ok := values[val]; ok {
			delete(values, val)
			continue
		}

		diff.Extra = append(diff.Extra, val)
	}

	var missing Collection
	for _, v := range values {
		missing = append(missing, v)
	}
	diff.Missing = missing

	return diff
}

func (e Collection) valueFrom(item reflect.Value) string {
	var val string

	switch item.Type().Kind() {
	case reflect.Struct:
		val = e.fieldValue(item)

		if val == "" {
			val = fmt.Sprintf("%#v", item.Interface())
		}
	default:
		val = fmt.Sprintf("%#v", item.Interface())
	}

	return val
}

func (e Collection) fieldValue(item reflect.Value) string {
	if len(e) == 0 {
		panic("Diff: collection is empty")
	}

	enum := e[0]
	typ := item.Type()

	if !strings.HasSuffix(enum.Type, typ.String()) {
		return ""
	}

	field, ok := typ.FieldByName(enum.FieldName)
	if !ok {
		return ""
	}

	if tag, ok := field.Tag.Lookup("enums"); !ok || tag != "identifier" {
		return ""
	}

	return fmt.Sprintf("%#v", item.FieldByName(enum.FieldName).Interface())
}
