package ss

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestStructEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("STRUCT_ENV_VERBOSE", "1")
	os.Setenv("TEST_STRING", "test_string")
	os.Setenv("TEST_DURATION", "1s")
	os.Setenv("TEST_UINT", "123")
	os.Setenv("TEST_INT", "-456")
	os.Setenv("TEST_BOOL", "1")
	os.Setenv("TEST_FLOAT", "78.9")

	// 确保在测试结束时清除环境变量
	defer os.Unsetenv("STRUCT_ENV_VERBOSE")
	defer os.Unsetenv("TEST_STRING")
	defer os.Unsetenv("TEST_DURATION")
	defer os.Unsetenv("TEST_UINT")
	defer os.Unsetenv("TEST_INT")
	defer os.Unsetenv("TEST_BOOL")
	defer os.Unsetenv("TEST_FLOAT")

	// 创建一个测试结构体实例
	type TestStruct struct {
		String   *string        `env:"TEST_STRING"`
		Uint     *uint64        `env:"TEST_UINT"`
		Duration *time.Duration `env:"TEST_DURATION"`
		Int      *int64         `env:"TEST_INT"`
		Bool     *bool          `env:"TEST_BOOL"`
		Float    *float64       `env:"TEST_FLOAT"`

		TestString    string
		TestString2   string
		TestString3   *string
		TestDuration  time.Duration
		TestDuration2 *time.Duration
		TestUint      uint64
		TestUint2     *uint64
		TestInt       int64
		TestInt2      *int64
		TestBool      bool
		TestBool2     *bool
		TestFloat     float64
		TestFloat2    *float64

		TestFloatIgnore float64 `env:"-"`

		Other time.Time
		name  string
	}

	testStruct := TestStruct{}

	// 调用 StructEnv 函数
	StructEnv(&testStruct)

	// 断言环境变量被正确解析
	assert.Equal(t, "test_string", *testStruct.String)
	assert.Equal(t, "test_string", testStruct.TestString)
	assert.Equal(t, time.Second, *testStruct.Duration)
	assert.Equal(t, time.Second, testStruct.TestDuration)
	assert.Equal(t, uint64(123), *testStruct.Uint)
	assert.Equal(t, uint64(123), testStruct.TestUint)
	assert.Equal(t, int64(-456), *testStruct.Int)
	assert.Equal(t, int64(-456), testStruct.TestInt)
	assert.Equal(t, true, *testStruct.Bool)
	assert.Equal(t, true, testStruct.TestBool)
	assert.Equal(t, 78.9, *testStruct.Float)
	assert.Equal(t, 78.9, testStruct.TestFloat)
	assert.Equal(t, 0.0, testStruct.TestFloatIgnore)
}

// 如果你的结构体字段使用了非导出的名称，或者有 "-" 标签，或者有自定义的类型，你需要添加额外的测试用例来处理这些情况。
