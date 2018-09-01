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
// only have interface or slice of interfaces as there parameters and must
// return an interface and an optional error.
func (p *provider) Add(item interface{}) error {
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
		p.producers++
	} else if itype.Kind() == reflect.Ptr && itype.Elem().Kind() == reflect.Interface {
		itype = itype.Elem()
		vitem = vitem.Elem()
	} else {
		return errors.New(
			"Invalid Item, interface prt or function returning interface required")
	}
	stype := reflect.SliceOf(itype)
	if v, ok := p.values[itype]; ok {
		s := make([]reflect.Value, 0, 5)
		//s := reflect.MakeSlice(valueSliceType, 0, 5)
		s = append(s, v)
		s = append(s, vitem)
		p.values[stype] = reflect.ValueOf(s)
		delete(p.values, itype)
	} else if s, ok := p.values[stype]; ok {
		p.values[stype] = reflect.Append(s, reflect.ValueOf(vitem))
	} else {
		p.values[itype] = vitem
	}
	return nil
}

// Build calls all added functions once.  If any functions return errors or
// any functions have dependencies that have not been added or there are any
// circular references a slice of errors will be returned.  If all goes well
// a Provider interface will be returned and the error slice result will be
// nil (not an empty slice!  would an empty slice be better?)
func (p *provider) Build() (Provider, []error) {
	prevProducers := p.producers + 1
	errs := make([]error, 0, 10)
	for p.producers > 0 && p.producers < prevProducers {
		prevProducers = p.producers
		errs = errs[:0]
		for itype, value := range p.values {
			kind := value.Kind()
			if kind == reflect.Slice && value.Type().Elem() == valueType {
				var errBad error
				errs, errBad = p.buildSlice(itype, value, errs)
				if errBad != nil {
					return nil, []error{errBad}
				}
			} else if kind == reflect.Func {
				cvalue, err, errBad := p.resolveProvider(value)
				if errBad != nil {
					return nil, []error{errBad}
				}
				if err != nil {
					errs = append(errs, err)
				} else {
					p.producers--
					p.values[itype] = cvalue
				}
			}
		}
	}
	if p.producers > 0 {
		return nil, errs
	}
	return p, nil
}

func (p *provider) buildSlice(itype reflect.Type, value reflect.Value, errs []error) ([]error, error) {
	s := value.Interface().([]reflect.Value)
	complete := true
	for i, svalue := range s {
		if svalue.Kind() == reflect.Func {
			cvalue, err, errBad := p.resolveProvider(svalue)
			if errBad != nil {
				return errs, errBad
			}
			if err != nil {
				errs = append(errs, err)
				complete = false
			} else {
				p.producers--
				s[i] = cvalue
			}
		}
	}
	if complete {
		cs := reflect.MakeSlice(itype, 0, len(s))
		for _, value = range s {
			cs = reflect.Append(cs, value)
		}
		p.values[itype] = cs
	}
	return errs, nil
}

func (p *provider) resolveProvider(f reflect.Value) (reflect.Value, error, error) {
	t := f.Type()
	in := make([]reflect.Value, t.NumIn())
	for i := 0; i < len(in); i++ {
		ptype := t.In(i)
		anIn, ok := p.values[ptype]
		if !ok {
			// check if we want slice but only one added
			if ptype.Kind() == reflect.Slice {
				anIn, ok = p.values[ptype.Elem()]
				if ok {
					if anIn.Kind() != reflect.Interface {
						return nilValue,
							fmt.Errorf("Can't resolve provider, input not built, type: %v", ptype),
							nil
					}
					in[i] = reflect.Append(reflect.MakeSlice(ptype, 0, 1), anIn)
					continue
				}
			}
			// bad will be no way to resolve this type ever
			return nilValue, nil, fmt.Errorf("No value of required type: %v", ptype)
		}
		if anIn.Kind() == reflect.Slice {
			if sliceIncomplete(anIn) {
				// slice contains unresolved provider(s) maybe will get resolved
				return nilValue,
					fmt.Errorf("Can't resolve provider, slice incomplete, type: %v", ptype),
					nil
			}
		} else if anIn.Kind() != reflect.Interface {
			return nilValue,
				fmt.Errorf("Can't resolve provider, input not built, type: %v", ptype),
				nil
		}
		in[i] = anIn
	}
	results := f.Call(in)
	if len(results) > 1 && !results[1].IsNil() {
		return nilValue, nil, results[1].Interface().(error)
	}
	if results[0].IsNil() {
		return nilValue,
			nil,
			fmt.Errorf("Provider returned nil value, type: %v", t.Out(0))
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

// ProvideFor calls the input function providing parameters it requires.  will
// return an error if any parameters can not be provided.  Otherwist the first
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
			return nil, fmt.Errorf("No value of required type: %v", ptype)
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
