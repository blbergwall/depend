package depend

import (
	"reflect"
	"testing"

	"github.com/blbergwall/assert"
)

func TestProducerOutputValid(t *testing.T) {
	a := assert.New(t)

	a.IsTrue(producerOutputValid(reflect.TypeOf(new1Result1)))
	a.IsTrue(producerOutputValid(reflect.TypeOf(new1AndErrorResult)))

	a.IsFalse(producerOutputValid(reflect.TypeOf(noResult)))
	a.IsFalse(producerOutputValid(reflect.TypeOf(nonInterfaceResult)))
}

func TestCanProvideFor(t *testing.T) {
	a := assert.New(t)

	a.IsTrue(canProvideFor(reflect.TypeOf(new1Result1)))
	a.IsTrue(canProvideFor(reflect.TypeOf(new2Consume1)))

	a.IsFalse(canProvideFor(reflect.TypeOf(intConsumer)))
}
