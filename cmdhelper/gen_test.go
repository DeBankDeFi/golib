package cmdhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testFlag struct {
	Name     string    `name:"name" type:"string" shorthand:"n" usage:"usage"`
	Array    []string  `name:"array" type:"string-slice" shorthand:"a" split:"," usage:"usage array"`
	IntArray []int     `name:"int-array" type:"int-slice" shorthand:"a" usage:"usage array"`
	Flag2    testFlag2 `name:"prefix"`
	Flag3    testFlag3
}

type testFlag2 struct {
	Foo  string `name:"foo" type:"string" enable-env:"true"`
	Bar  string `name:"bar" type:"string" enable-env:"false"`
	Bar2 string `name:"bar2" type:"string"`
}

type testFlag3 struct {
	Bar3 string `type:"string"`
}

func Test_resolveFlags(t *testing.T) {
	f := testFlag{
		Name:     "name",
		Array:    []string{"a"},
		IntArray: []int{80},
	}

	var result Flags
	result = resolveFlags(&f, result, "")
	assert.Equal(t, Flags{
		{Name: "name", FullName: "name", Shorthand: "n", Usage: "usage", Type: "string", Value: "name", Pointer: &f.Name},
		{Name: "array", FullName: "array", Split: ",", Shorthand: "a", Usage: "usage array", Type: "string-slice", Value: []string{"a"}, Pointer: &f.Array},
		{Name: "int-array", FullName: "int-array", Shorthand: "a", Usage: "usage array", Type: "int-slice", Value: []int{80}, Pointer: &f.IntArray},
		{Name: "foo", FullName: "prefix-foo", Type: "string", EnableEnv: true, FullEnv: "PREFIX_FOO", Value: "", Pointer: &f.Flag2.Foo},
		{Name: "bar", FullName: "prefix-bar", Type: "string", EnableEnv: false, Value: "", Pointer: &f.Flag2.Bar},
		{Name: "bar2", FullName: "prefix-bar2", Type: "string", Value: "", Pointer: &f.Flag2.Bar2},
		{Name: "bar3", FullName: "flag3-bar3", Type: "string", Value: "", Pointer: &f.Flag3.Bar3},
	}, result)
}

func Test_toSnake(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "InfluxDB",
			expected: "influx-db",
		},
		{
			name:     "InfluxDBV2",
			expected: "influx-dbv2",
		},
		{
			name:     "fooBarVer",
			expected: "foo-bar-ver",
		},
		{
			name:     "EtcdEndpoints",
			expected: "etcd-endpoints",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toSnake(tc.name)
			assert.Equal(t, tc.expected, result)
		})
	}
}
