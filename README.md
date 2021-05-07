enums
=====

Uses the AST to find all variables of a certain type in a package. 
This is useful for cases where you need to ensure that all values 
of a type is accounted for somewhere. There is an overlap with the
[exhaustive] project which does this for `switch` statements.

[exhaustive]: https://github.com/nishanths/exhaustive

For the sake of this package an `enum` is a named type with multiple
values which you're operating on as a collection. They could be strings,
ints, and even complex types like functions or structs.

## Example

For example, I have an enum type for feature flags. For each request
I call a 3rdparty service to check the state of each flag, so whenever
a new flag is created it needs to be added to the All function.

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

Using enums we can test that we haven't missed out on a new flag:

```golang
func TestAllFlagsCovered(t *testing.T) {
    enumstest.NoDiff(t, "./feature", "feature.Flag", full.AllFlags())
}
```

## License

See the [LICENSE](LICENSE.txt) file for license rights and limitations (MIT).
