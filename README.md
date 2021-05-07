enums
=====

Uses the AST to find all variables of a certain type in a package. 
This is useful for cases where you need to ensure that all values 
of a type is accounted for somewhere.

For example, I have a list of all feature flags and want to ensure
that any new flag is added to the list of all flags which is used to
fetch the state of the flag from a 3rdparty service.

## Example

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
    matches, err := enums.All("./feature", "feature.Flag")
    require.NoError(t, err)

    require.Equal(
        t,
        enums.Diff{
            Missing: enums.Collection{
                {
                    Type:  "cool.tld/feature.Flag",
                    Name:  "DeployOneThing",
                    Value: `"deploy-one-thing"`,
                },
            }
        },
        matches.Diff(full.AllFlags()),
        "expected to have an extra value indicated",
    )
}
```
