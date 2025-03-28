package falsepositiveinfunc

type Flag string

const (
	DeployAllTheThings Flag = "deploy-all-the-things"
)

var AnotherExample Flag = "hello-there"

// SlicedExample should be ignored, because I can't think of a case where this is what I'd like to match when it's declared top-level.
var SlicedExample = []Flag{"hello", "there"}

// FalsePositive is used to show a case where the flag is picked up as a declaration when it shouldn't be since it's not done at the file level.
func FalsePositive() []Flag {
	var myVar []Flag

	return myVar
}

type example struct{}

func (e *example) Method() Flag {
	return "m000"
}
