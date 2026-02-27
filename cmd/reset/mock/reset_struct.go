package mock

type ChildWithReset struct {
	Field string
}

func (cr *ChildWithReset) Reset() {
	cr.Field = ""
}

// generate:reset
type ResetableStruct struct {
	i              int
	str            string
	strP           *string
	intP           *int
	boolP          *bool
	s              []int
	m              map[string]string
	child          *ResetableStruct
	childWithReset *ChildWithReset
}

//// generate:reset
//type ResetableStruct2 struct {
//	i              int
//	str            string
//	strP           *string
//	intP           *int
//	boolP          *bool
//	s              []int
//	m              map[string]string
//	child          *ResetableStruct
//	childWithReset *ChildWithReset
//}
