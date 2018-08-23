// types and functions to be used by tests
package depend

import (
	"errors"
	"fmt"
)

type (
	testInterface1 interface {
		Method1() string
	}

	testInterface2 interface {
		Method2() string
	}

	testInterface3 interface {
		Method3() string
	}

	testStruct1 struct {
		field string
	}

	testStruct2 struct {
		field string
	}

	testStruct3 struct {
		field string
	}
)

func new1Result1() testInterface1 {
	return &testStruct1{field: "testInterface1 #1"}
}

func new1Result2() testInterface1 {
	return &testStruct1{field: "testInterface1 #2"}
}

func new1Result3() testInterface1 {
	return &testStruct1{field: "testInterface1 #3"}
}

func new1AndErrorResult() (testInterface1, error) {
	return &testStruct1{field: "testInterface1 with no error"}, nil
}

func new1AndErrorError() (testInterface1, error) {
	return nil, errors.New("new1AndErrorError testing error")
}

func new1Consume2(i testInterface2) testInterface1 {
	return &testStruct1{field: "testInterface1 #1"}
}

func new1Nil() testInterface1 {
	return nil
}

func noResult() {
	return
}

func nonInterfaceResult() int {
	return 1
}

func new2Consume1(i1 testInterface1) (testInterface2, error) {
	return &testStruct2{field: fmt.Sprint("testInterface2 from ", i1.Method1())}, nil
}

func new2ConsumeSlice1(i1s []testInterface1) (testInterface2, error) {
	return &testStruct2{field: "testInterface2 from []testInterface1"}, nil
}

func new3Consume2(i testInterface2) (testInterface3, error) {
	return &testStruct3{field: "testInterface3 from testInterface2"}, nil
}

func new3ConsumeSlice2(i []testInterface2) (testInterface3, error) {
	return &testStruct3{field: "testInterface3 from []testInterface2"}, nil
}

func intConsumer(i int) (testInterface1, error) {
	return &testStruct1{field: "Hi"}, nil
}

func (ts1 *testStruct1) Method1() string {
	return ts1.field
}

func (ts2 *testStruct2) Method2() string {
	return ts2.field
}

func (ts3 *testStruct3) Method3() string {
	return ts3.field
}

func toProvideFor(i1s []testInterface1, i2 testInterface2) (string, error) {
	return i1s[0].Method1(), nil
}

func toProvideForNoReturn(i1 testInterface1) {
}
