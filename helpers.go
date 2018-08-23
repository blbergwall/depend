package depend

import (
	"reflect"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

func producerOutputValid(itype reflect.Type) bool {
	ocount := itype.NumOut()
	return ocount >= 1 && ocount <= 2 &&
		itype.Out(0).Kind() == reflect.Interface &&
		(ocount < 2 || itype.Out(1) == errorType)
}

func canProvideFor(itype reflect.Type) bool {
	for icount := itype.NumIn() - 1; icount >= 0; icount-- {
		if !canProvide(itype.In(icount)) {
			return false
		}
	}
	return true
}

func canProvide(itype reflect.Type) bool {
	kind := itype.Kind()
	return kind == reflect.Interface ||
		(kind == reflect.Slice && itype.Elem().Kind() == reflect.Interface)
	// maybe also struct pointer where all interfaces in struct would be provided?
}
