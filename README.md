typedecl
=====

[![godocs.io](http://godocs.io/github.com/gaqzi/typedecl.svg)](http://godocs.io/github.com/gaqzi/typedecl)?status.svg)](http://godocs.io/github.com/gaqzi/typedecl)

Uses the AST to find all variables of a certain type in a package. 
This is useful for cases where you need to ensure that all values 
of a type is accounted for somewhere. There is an overlap with the
[exhaustive] project which does this for `switch` statements.

This package finds all top-level declarations of a named type that are
[basic literals] (strings, numbers) and structs,
so that you can perform logic on all of those declared values.

[exhaustive]: https://github.com/nishanths/exhaustive
[basic literals]: https://go.dev/ref/spec#Operands

## Example

For example, I have a type for feature flags. For each request
I call a 3rdparty service to check the state of each flag, so whenever
a new flag is created it needs to be added to the `AllFlags` function.

```golang
type Flag string

const (
    DeployAllTheThings Flag = "deploy-all-the-things"
    // This flag is new and we need to remember to update AllFlags below
    DeployOneThing     Flag = "deploy-one-thing"
)

func AllFlags() []Flag {
    return []Flag{
        DeployAllTheThings,
    }
}
```

Using `typedecl` we can test that we haven't missed out on a new flag:

```golang
func TestAllFlagsCovered(t *testing.T) {
    typedecltest.NoDiff(t, "./feature", "feature.Flag", full.AllFlags())
}
```

## Using with structs

We need a way to uniquely identify values in a struct, so the identifier 
will still have to be a basic literal, and this is done with the 
`typedecl:"identifier"` tag.

```golang
type DefaultFlag struct {
    Name string `typedecl:"identifier"`
    IsOn bool
}
```

## License

See the [LICENSE](LICENSE.txt) file for license rights and limitations (MIT).
