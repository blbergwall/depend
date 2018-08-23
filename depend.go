// Package depend is a simple dependency injection tool for go.
package depend

import (
	"errors"
	"fmt"
	"reflect"
)

type (
	// Builder allows assing all dependencies and then building the Provider
	Builder interface {
		Add(item interface{}) error

		Build() (Provider, []error)
	}

	// Provider runs functions providing an input dependencies that may be
	// needed.
	Provider interface {
		ProvideFor(f interface{}) (interface{}, error)
	}

	provider struct {
		producers int

		values map[reflect.Type]reflect.Value
	}
)

var (
	nilValue  = reflect.ValueOf(nil)
	valueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()
)

// New creats a Builder
func New() Builder {
	return &provider{
		producers: 0,
		values:    make(map[reflect.Type]reflect.Value),
	}
}

// Add takes interface *pointers* or functions that return an interface so
// those interfaces can be provided to other such functions.  Functions may
// only have interface or slice of interfaces as there paramiters and must
// return an interface and an optional error.
func (b *provider) Add(item interface{}) error {
	itype := reflect.TypeOf(item)
	if itype == nil {
		return errors.New("can't add nil")
	}
	vitem := reflect.ValueOf(item)
	kind := itype.Kind()
	if kind == reflect.Func {
		if !producerOutputValid(itype) {
			return errors.New("Producer function must return 1 interface and an optional error")
		}
		if !canProvideFor(itype) {
			return errors.New("Producer inputs must be interface or slice of interfaces")
		}
		itype = itype.Out(0)
		b.producers++
	} else if itype.Kind() == reflect.Ptr && itype.Elem().Kind() == reflect.Interface {
		itype = itype.Elem()
		vitem = vitem.Elem()
	} else {
		return errors.New(
			"Invalid Item, interface prt or function returning interface required")
	}
	stype := reflect.SliceOf(itype)
	if v, ok := b.values[itype]; ok {
		s := make([]reflect.Value, 0, 5)
		//s := reflect.MakeSlice(valueSliceType, 0, 5)
		s = append(s, v)
		s = append(s, vitem)
		b.values[stype] = reflect.ValueOf(s)
		delete(b.values, itype)
	} else if s, ok := b.values[stype]; ok {
		b.values[stype] = reflect.Append(s, reflect.ValueOf(vitem))
	} else {
		b.values[itype] = vitem
	}
	return nil
}

// Build calls all added functions once.  If any functions return errors or
// any functions have dependencies that have not been added or there are any
// circular references a slice of errors will be returned.  If all goes well
// a Provider interface will be returned and the error slice result will be
// nil (not an empty slice!  would an empty slice be better?)
func (b *provider) Build() (Provider, []error) {
	prevProducers := b.producers + 1
	errs := make([]error, 0, 10)
	for b.producers > 0 && b.producers < prevProducers {
		prevProducers = b.producers
		errs = errs[:0]
		for itype, value := range b.values {
			kind := value.Kind()
			if kind == reflect.Slice && value.Type().Elem() == valueType {
				s := value.Interface().([]reflect.Value)
				complete := true
				for i, svalue := range s {
					if svalue.Kind() == reflect.Func {
						cvalue, err, errBad := b.resolveProvider(svalue)
						if errBad != nil {
							return nil, []error{errBad}
						}
						if err != nil {
							errs = append(errs, err)
							complete = false
						} else {
							b.producers--
							s[i] = cvalue
						}
					}
				}
				if complete {
					cs := reflect.MakeSlice(itype, 0, len(s))
					for _, value = range s {
						cs = reflect.Append(cs, value)
					}
					b.values[itype] = cs
				}
			} else if kind == reflect.Func {
				cvalue, err, errBad := b.resolveProvider(value)
				if errBad != nil {
					return nil, []error{errBad}
				}
				if err != nil {
					errs = append(errs, err)
				} else {
					b.producers--
					b.values[itype] = cvalue
				}
			}
		}
	}
	if b.producers > 0 {
		return nil, errs
	}
	return b, nil
}

func (b *provider) resolveProvider(f reflect.Value) (reflect.Value, error, error) {
	t := f.Type()
	in := make([]reflect.Value, t.NumIn())
	for i := 0; i < len(in); i++ {
		ptype := t.In(i)
		anIn, ok := b.values[ptype]
		if !ok {
			// check if we want slice but only one added
			if ptype.Kind() == reflect.Slice {
				anIn, ok = b.values[ptype.Elem()]
				if ok {
					if anIn.Kind() != reflect.Interface {
						return nilValue,
							errors.New(
								fmt.Sprintf("Can't resolve provider, input not built, type: %v", ptype),
							),
							nil
					}
					in[i] = reflect.Append(reflect.MakeSlice(ptype, 0, 1), anIn)
					continue
				}
			}
			// bad will be no way to resolve this type ever
			return nilValue, nil, errors.New(fmt.Sprintf("No value of required type: %v", ptype))
		}
		if anIn.Kind() == reflect.Slice {
			if sliceIncomplete(anIn) {
				// slice contains unresolved provider(s) maybe will get resolved
				return nilValue,
					errors.New(
						fmt.Sprintf("Can't resolve provider, slice incomplete, type: %v", ptype),
					),
					nil
			}
		} else if anIn.Kind() != reflect.Interface {
			return nilValue,
				errors.New(
					fmt.Sprintf("Can't resolve provider, input not built, type: %v", ptype),
				),
				nil
		}
		in[i] = anIn
	}
	results := f.Call(in)
	if len(results) > 1 && !results[1].IsNil() {
		return nilValue, nil, results[1].Interface().(error)
	}
	if results[0].IsNil() {
		return nilValue, nil, errors.New(
			fmt.Sprintf("Provider returned nil value, type: %v", t.Out(0)),
		)
	}
	return results[0], nil, nil
}

func sliceIncomplete(v reflect.Value) bool {
	for i := v.Len() - 1; i >= 0; i-- {
		if v.Index(i).Kind() != reflect.Interface {
			return true
		}
	}
	return false
}

// ProvideFor calls the input fucntion providing paramiters it requires.  will
// return an error if any paramiters can not be provided.  Otherwist the first
// result of the function is returned as the first result and if the last
// result of the function is of type error it is rturned as the second (and
// final) result.  If the function only returnes an error these will be
// the same.  If the function returns more than 2 results (or the last result
// is not of type error) extra results will be discarded.
func (p *provider) ProvideFor(f interface{}) (interface{}, error) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return nil, errors.New("Can only ProvideFor function")
	}
	in := make([]reflect.Value, t.NumIn())
	for i := 0; i < len(in); i++ {
		ptype := t.In(i)
		anIn, ok := p.values[ptype]
		if !ok {
			// check if we want slice but only one added
			if ptype.Kind() == reflect.Slice {
				anIn, ok = p.values[ptype.Elem()]
				if ok {
					in[i] = reflect.Append(reflect.MakeSlice(ptype, 0, 1), anIn)
					continue
				}
			}
			return nil, errors.New(fmt.Sprintf("No value of required type: %v", ptype))
		}
		in[i] = anIn
	}
	results := reflect.ValueOf(f).Call(in)
	resultsCount := len(results)
	if resultsCount > 0 {
		err, _ := results[resultsCount-1].Interface().(error)
		result := results[0].Interface()
		return result, err
	}
	return nil, nil
}