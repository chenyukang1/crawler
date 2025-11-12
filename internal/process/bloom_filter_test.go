package process

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	tests := []struct {
		name string
		size uint
		adds []string
		arg  string
		want bool
	}{
		{
			"test contains",
			8,
			[]string{"1", "2", "3"},
			"1",
			true,
		},
		{
			"test base64 contains",
			1024,
			[]string{"MTc2Mjc3MDM4MAo=", "MTc2Mjc3MDM4MDExMTExMTExMTExMTExMTExMTExMQo="},
			"MTc2Mjc3MDM4MAo=",
			true,
		},
		{
			"test not contains",
			1024,
			[]string{"MTc2Mjc3MDM4MAo=", "MTc2Mjc3MDM4MDExMTExMTExMTExMTExMTExMTExMQo="},
			"MTc2Mjc3MDM4MAo",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewBloomFilter(tt.size)
			for _, add := range tt.adds {
				f.Add(add)
			}
			if got := f.Contains(tt.arg); got != tt.want {
				t.Errorf("Contains(%s) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func Test_djb2Hash(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{
			"test consistency",
			"MTc2Mjc3MDM4MAo=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := djb2Hash(tt.arg)
			second := djb2Hash(tt.arg)
			if first != second {
				t.Errorf("first %d != second %d", first, second)
			}
		})
	}
}

func Test_fnvHash(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{
			"test consistency",
			"MTc2Mjc3MDM4MAo=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := fnvHash(tt.arg)
			second := fnvHash(tt.arg)
			if first != second {
				t.Errorf("first %d != second %d", first, second)
			}
		})
	}
}
