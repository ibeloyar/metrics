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
	//obj0 := TestObject{3}
	//obj0.Reset()
	//if obj0.value != 0 {
	//	t.Errorf("zero get with value %d", obj0.value)
	//}
	pool := New[*TestObject]()

	if pool.Len() != 0 {
		t.Errorf("pool len wust be 0 got %d", pool.Len())
	}
	obj1 := pool.Get()
	if obj1 != nil {
		t.Errorf("first pool get must be nil, got %v", obj1)
	}

	testObj1 := TestObject{value: 1}
	testObj2 := TestObject{value: 2}

	pool.Put(&testObj1)
	pool.Put(&testObj2)

	if pool.Len() != 2 {
		t.Errorf("pool len wust be 2 got %d", pool.Len())
	}

	obj2 := pool.Get()
	if obj2 == nil {
		t.Errorf("second pool get must be non-nil")
	}
	if obj2.value != 0 {
		t.Errorf("second pool get must be zero got %d", obj2.value)
	}
	if pool.Len() != 1 {
		t.Errorf("pool len wust be 1 got %d", pool.Len())
	}
}
