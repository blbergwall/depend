package depend

import (
	"reflect"
	"testing"

	"github.com/blbergwall/assert"
)

func TestProducerOutputValid(t *testing.T) {
	a := assert.New(t)

	a.True(producerOutputValid(reflect.TypeOf(new1Result1)))
	a.True(producerOutputValid(reflect.TypeOf(new1AndErrorResult)))

	a.False(producerOutputValid(reflect.TypeOf(noResult)))
	a.False(producerOutputValid(reflect.TypeOf(nonInterfaceResult)))
}

func TestCanProvideFor(t *testing.T) {
	a := assert.New(t)

	a.True(canProvideFor(reflect.TypeOf(new1Result1)))
	a.True(canProvideFor(reflect.TypeOf(new2Consume1)))

	a.False(canProvideFor(reflect.TypeOf(intConsumer)))
}
