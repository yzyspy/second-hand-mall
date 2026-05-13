package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSet 测试 Set 结构体
func TestSet(t *testing.T) {
	set := &Set{}
	set.Add("a")
	set.Add("b")
	set.Add("c")

	assert.Equal(t, 3, set.Len())
	assert.True(t, set.Contains("a"))
	assert.True(t, set.Contains("b"))
	assert.True(t, set.Contains("c"))

	//遍历set
	for key, item := range *set {
		t.Log(key, item)
	}
	for i := 0; i < set.Len(); i++ {
		t.Log(set)
	}
}
