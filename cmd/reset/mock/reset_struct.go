package mock

type ChildWithReset struct {
	Field string
}

func (cr *ChildWithReset) Reset() {
	cr.Field = ""
}

// generate:reset
type ResettableStruct struct {
	i              int
	str            string
	strP           *string
	intP           *int
	boolP          *bool
	s              []int
	m              map[string]string
	child          *ResettableStruct
	childWithReset *ChildWithReset
}

type NonResettableStruct struct {
	resettable bool
}

// generate:reset
type ResettableStruct2 struct {
	bool           bool
	float32P       *float32
	float64P       *float64
	m              map[int]string
	child          *ResettableStruct
	childWithReset *ChildWithReset
}
