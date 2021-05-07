package full

type Flag string

const (
	DeployAllTheThings Flag = "deploy-all-the-things"
	DeployOneThing     Flag = "deploy-one-thing"
)

// AllFlags is used to show the perfect scenario, the flags and the code are in harmony
func AllFlags() []Flag {
	return []Flag{
		DeployAllTheThings,
		DeployOneThing,
	}
}

// ExtraFlags a bit of a contrived example of returning more data than what's expected
func ExtraFlags() []interface{} {
	return []interface{}{
		DeployAllTheThings,
		DeployOneThing,
		"m000",
	}
}

// MissingFlags returns all flags except one to show what it looks like when you've forgotten to add it
func MissingFlags() []Flag {
	return []Flag{DeployAllTheThings}
}
