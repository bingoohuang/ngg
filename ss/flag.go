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
