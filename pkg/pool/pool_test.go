package pool

import (
	"testing"
)

type TestObject struct {
	value int
}

func (t *TestObject) Reset() {
	t.value = 0
}

func TestPoolBasic(t *testing.T) {
	pool := New[*TestObject]()

	obj1 := pool.Get()
	if obj1 != nil {
		t.Errorf("first pool get must be nil, got %v", obj1)
	}

	testObj1 := TestObject{value: 1}
	testObj2 := TestObject{value: 2}

	pool.Put(&testObj1)
	pool.Put(&testObj2)

	obj2 := pool.Get()
	if obj2 == nil {
		t.Errorf("second pool get must be non-nil")
	}
	if obj2 != nil && obj2.value != 0 {
		t.Errorf("second pool get must be zero got %d", obj2.value)
	}
}
