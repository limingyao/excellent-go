package stringx

import "testing"

func TestJoin(t *testing.T) {
	t.Log(Join([]float32{1.1, 2.2, 3.3}, ","))
	t.Log(Join([]float32{1.1, -2.2, 3.3}, ","))
	t.Log(Join([]int{1, 2, 3}, ","))
	t.Log(Join([]int{1, -2, 3}, ","))
}
