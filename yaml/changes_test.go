package yaml_test

import (
	"bytes"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

// 嵌套结构体
type Address struct {
	Street     string
	City       string
	PostalCode int
}

// 主结构体
type Person struct {
	Name       string
	Age        int
	IsStudent  bool
	Height     float64
	Address    Address
	Children   []string
	Grades     map[string]int
	SecretKey  *string // 指针类型
	EmptyField any     // 空接口
}

const expectedKeyMatchMode = `Name: Alice
Age: 30
IsStudent: false
Height: 1.75
Address:
  Street: 123 Main St
  City: New York
  PostalCode: 10001
Children:
- Bob
- Charlie
Grades:
  Math: 90
  Science: 85
SecretKey: abc123
EmptyField: null
`

func TestKeyMatchMode(t *testing.T) {
	key := "abc123"
	p := Person{
		Name:      "Alice",
		Age:       30,
		IsStudent: false,
		Height:    1.75,
		Address: Address{
			Street:     "123 Main St",
			City:       "New York",
			PostalCode: 10001,
		},
		Children:  []string{"Bob", "Charlie"},
		Grades:    map[string]int{"Math": 90, "Science": 85},
		SecretKey: &key,
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf,
		// 序列化时，key 严格匹配结构体字段名称（在没有定义 yaml 标签名称时）
		yaml.WithEncodeKeyMatchMode(yaml.KeyMatchStrict),
	)
	encoder.Encode(p)
	assert.Equal(t, expectedKeyMatchMode, buf.String())

	decoder := yaml.NewDecoder(bytes.NewReader(buf.Bytes()),
		// 反序列化时，key 严格匹配结构体字段名称（在没有定义 yaml 标签名称时）
		yaml.WithDecodeKeyMatchMode(yaml.KeyMatchStrict),
	)
	var p2 Person
	decoder.Decode(&p2)
	assert.Equal(t, p, p2)
}
