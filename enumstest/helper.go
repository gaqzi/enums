package enumstest

import (
	"github.com/gaqzi/enums"
)

type tHelper interface {
	Helper()
	Log(...interface{})
	Fail()
}

// NoDiff looks up all types in pkg and asserts they they have all the values from actual
//
// Example:
//
//	NoDiff(t, "./feature", "feature.Flag", []feature.Flag{"flag1", "flag2"})
func NoDiff(t tHelper, pkg, typ string, actual interface{}, failureMsg ...string) bool {
	t.Helper()

	collection, err := enums.All(pkg, typ)
	if err != nil {
		t.Log("failed to load enums.All: " + err.Error())
		t.Fail()
		return false
	}

	diff := collection.Diff(actual)
	if diff.Zero() {
		return true
	}

	var msg string
	if len(failureMsg) > 0 {
		msg += failureMsg[0] + "\n"
	}

	t.Log(msg + diff.String())
	t.Fail()
	return false
}
