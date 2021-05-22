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

// MissingFlags returns all flags except one to show what it looks like when you've forgotten to add it
func MissingFlags() []Flag {
	return []Flag{DeployAllTheThings}
}
