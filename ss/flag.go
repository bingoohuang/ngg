package ss

type FlagStringBool struct {
	Val    string
	Exists bool
}

func (i *FlagStringBool) Type() string     { return "stringbool" }
func (i *FlagStringBool) String() string   { return i.Val }
func (i *FlagStringBool) Get() interface{} { return i.Val }
func (i *FlagStringBool) Set(value string) error {
	i.Val = value
	i.Exists = true
	return nil
}

func (i *FlagStringBool) SetExists(b bool) { i.Exists = b }

func NewFlagSize(up *uint64, val string) *FlagSize {
	if val != "" {
		*up, _ = ParseBytes(val)
	}
	return &FlagSize{Val: up}
}

type FlagSize struct {
	Val *uint64
}

func (i *FlagSize) Type() string { return "size" }

func (i *FlagSize) String() string {
	if i.Val == nil {
		return "0"
	}
	return Bytes(*i.Val)
}

func (i *FlagSize) Set(value string) (err error) {
	*i.Val, err = ParseBytes(value)
	return err
}
