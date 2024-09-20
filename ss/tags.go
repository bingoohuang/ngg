package ss

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/*

https://github.com/fatih/structtag

structtag provides a way of parsing and manipulating struct tag Go fields.
It's used by tools like gomodifytags. For more examples, checkout the projects using structtag.

type t struct {
		t string `json:"foo,omitempty,string" xml:"foo"`
	}

	// get field tag
	tag := reflect.TypeOf(t{}).Field(0).Tag

	// ... and start using structtag by parsing the tag
	tags, err := ss.ParseStructTags(string(tag))
	if err != nil {
		panic(err)
	}

	// iterate over all tags
	for _, t := range tags.Tags() {
		fmt.Printf("tag: %+v\n", t)
	}

	// get a single tag
	jsonTag, err := tags.Get("json")
*/

var (
	ErrTagSyntax      = errors.New("bad syntax for struct tag pair")
	ErrTagKeySyntax   = errors.New("bad syntax for struct tag key")
	ErrTagValueSyntax = errors.New("bad syntax for struct tag value")
	ErrKeyNotExist    = errors.New("tag key does not exist")
	ErrTagNotExist    = errors.New("tag does not exist")
)

// StructTags represent a set of tags from a single struct field
type StructTags struct {
	tags    []*StructTag
	tagsMap map[string]*StructTag
}

// StructTag defines a single struct's string literal tag
type StructTag struct {
	// Key is the tag key, such as json, xml, etc..
	// i.e: `json:"foo,omitempty". Here key is: "json"
	Key string

	// Name is a part of the value
	// i.e: `json:"foo,omitempty". Here name is: "foo"
	Name string

	// Options is a part of the value. It contains a slice of tag options i.e:
	// `json:"foo,omitempty". Here options is: ["omitempty"]
	Options []string

	// OptionsMap is a map of the options
	// i.e: `json:"foo,omitempty". Here OptionsMap is: {"foo":"", "omitempty": ""}
	OptionsMap map[string]string

	// Raw is the raw value of the tag
	// i.e: `json:"foo,omitempty". Here Raw is: "foo,omitempty"
	Raw string
}

// Parse parses a single struct field tag and returns the set of tags.
func ParseStructTags(tag string) (*StructTags, error) {
	var tags []*StructTag

	hasTag := tag != ""

	// NOTE(arslan) following code is from reflect and vet package with some
	// modifications to collect all necessary information and extend it with
	// usable methods
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax
		// error. Strictly speaking, control chars include the range [0x7f,
		// 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
		// multi-byte control characters as it is simpler to inspect the tag's
		// bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 {
			return nil, ErrTagKeySyntax
		}
		if i+1 >= len(tag) || tag[i] != ':' {
			return nil, ErrTagSyntax
		}
		if tag[i+1] != '"' {
			return nil, ErrTagValueSyntax
		}

		key := tag[:i]
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, ErrTagValueSyntax
		}

		qvalue := tag[:i+1]
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, ErrTagValueSyntax
		}

		structTag := ParseStructTag(key, value)
		tags = append(tags, structTag)
	}

	if hasTag && len(tags) == 0 {
		return nil, nil
	}

	tagsMap := make(map[string]*StructTag)
	for _, tag := range tags {
		tagsMap[tag.Key] = tag
	}

	return &StructTags{
		tags:    tags,
		tagsMap: tagsMap,
	}, nil
}

func ParseStructTag(key, value string) *StructTag {
	res := Split(value, ",")
	name := ""
	if len(res) > 0 {
		name = res[0]
	}

	var options []string
	if len(res) > 1 {
		options = res[1:]
	}

	optionsMap := SplitToMap(value, ",", "=")
	delete(optionsMap, "-")
	if len(optionsMap) == 0 {
		optionsMap = nil
	}

	return &StructTag{
		Key:        key,
		Name:       name,
		Options:    options,
		Raw:        value,
		OptionsMap: optionsMap,
	}
}

// Get returns the tag associated with the given key. If the key is present
// in the tag the value (which may be empty) is returned. Otherwise, the
// returned value will be the empty string. The ok return value reports whether
// the tag exists or not (which the return value is nil).
func (t *StructTags) Get(key string) (*StructTag, error) {
	if tag, ok := t.tagsMap[key]; ok {
		return tag, nil
	}

	return nil, ErrTagNotExist
}

func (t *StructTags) GetTag(key string) string {
	if tag, ok := t.tagsMap[key]; ok {
		return tag.Raw
	}

	return ""
}

// GetOpt gets opt's value by its name
func (t *StructTag) GetOpt(optName string) string {
	if opt, ok := t.OptionsMap[optName]; ok && opt != "" {
		return opt
	}

	return ""
}

// Set sets the given tag. If the tag key already exists it'll override it
func (t *StructTags) Set(tag *StructTag) error {
	if tag.Key == "" {
		return ErrKeyNotExist
	}

	added := false
	for i, tg := range t.tags {
		if tg.Key == tag.Key {
			added = true
			t.tags[i] = tag
		}
	}

	if !added {
		// this means this is a new tag, add it
		t.tags = append(t.tags, tag)
	}

	t.tagsMap[tag.Key] = tag
	return nil
}

// AddOptions adds the given option for the given key. If the option already
// exists it doesn't add it again.
func (t *StructTags) AddOptions(key string, options ...string) error {
	tag, err := t.Get(key)
	if err != nil {
		return err
	}

	for _, opt := range options {
		if !tag.HasOption(opt) {
			tag.Options = append(tag.Options, opt)
		}
	}

	return nil
}

// DeleteOptions deletes the given options for the given key
func (t *StructTags) DeleteOptions(key string, options ...string) error {
	tag, err := t.Get(key)
	if err != nil {
		return err
	}

	hasOption := func(option string) bool {
		for _, opt := range options {
			if opt == option {
				return true
			}
		}
		return false
	}

	var updated []string
	for _, opt := range tag.Options {
		if !hasOption(opt) {
			updated = append(updated, opt)
		}
	}

	tag.Options = updated
	return nil
}

// Delete deletes the tag for the given keys
func (t *StructTags) Delete(keys ...string) {
	hasKey := func(key string) bool {
		for _, k := range keys {
			if k == key {
				return true
			}
		}
		return false
	}

	var updated []*StructTag
	for _, tag := range t.tags {
		if !hasKey(tag.Key) {
			updated = append(updated, tag)
		}
	}
	for _, k := range keys {
		delete(t.tagsMap, k)
	}

	t.tags = updated
}

// Tags returns a slice of tags. The order is the original tag order unless it
// was changed.
func (t *StructTags) Tags() []*StructTag {
	return t.tags
}

// Keys returns a slice of tags' keys.
func (t *StructTags) Keys() []string {
	var keys []string
	for _, tag := range t.tags {
		keys = append(keys, tag.Key)
	}
	return keys
}

// String reassembles the tags into a valid literal tag field representation
func (t *StructTags) String() string {
	tags := t.Tags()
	if len(tags) == 0 {
		return ""
	}

	var buf bytes.Buffer
	for i, tag := range t.Tags() {
		buf.WriteString(tag.String())
		if i != len(tags)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

// HasOption returns true if the given option is available in options
func (t *StructTag) HasOption(opt string) bool {
	for _, tagOpt := range t.Options {
		if tagOpt == opt {
			return true
		}
	}

	return false
}

// Value returns the raw value of the tag, i.e. if the tag is
// `json:"foo,omitempty", the Value is "foo,omitempty"
func (t *StructTag) Value() string {
	options := strings.Join(t.Options, ",")
	if options != "" {
		return fmt.Sprintf(`%s,%s`, t.Name, options)
	}
	return t.Name
}

// String reassembles the tag into a valid tag field representation
func (t *StructTag) String() string {
	return fmt.Sprintf(`%s:%q`, t.Key, t.Value())
}

// GoString implements the fmt.GoStringer interface
func (t *StructTag) GoString() string {
	template := `{
		Key:    '%s',
		Name:   '%s',
		Option: '%s',
	}`

	if t.Options == nil {
		return fmt.Sprintf(template, t.Key, t.Name, "nil")
	}

	options := strings.Join(t.Options, ",")
	return fmt.Sprintf(template, t.Key, t.Name, options)
}

func (t *StructTags) Len() int {
	return len(t.tags)
}

func (t *StructTags) Less(i int, j int) bool {
	return t.tags[i].Key < t.tags[j].Key
}

func (t *StructTags) Swap(i int, j int) {
	t.tags[i], t.tags[j] = t.tags[j], t.tags[i]
}
