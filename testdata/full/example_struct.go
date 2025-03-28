package full

type FlagStruct struct {
	Name      string `typedecl:"identifier"`
	DefaultOn bool
}

var (
	FlagDefaultOn = FlagStruct{Name: "flag-default-on", DefaultOn: true}
)

func AllFlagStruct() []FlagStruct {
	return []FlagStruct{FlagDefaultOn}
}

func MissingFlagStruct() []FlagStruct {
	return []FlagStruct{}
}
