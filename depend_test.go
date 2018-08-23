package depend

import (
	"testing"

	"github.com/blbergwall/assert"
)

func TestNewAndEmptyBuild(t *testing.T) {
	a := assert.New(t)

	b := New()
	a.NotNil(b)
	p, err := b.Build()
	a.AreEqual(0, len(err))
	a.NotNil(p)
}

func TestAddErrors(t *testing.T) {
	a := assert.New(t)

	b := New()
	a.ErrorContains(b.Add(nil), "can't add nil")
	iface1_1 := new1Result1()
	a.ErrorContains(b.Add(iface1_1), "Invalid Item, interface prt")
	a.ErrorContains(b.Add(nonInterfaceResult), "Producer function must return")
	a.ErrorContains(b.Add(intConsumer), "Producer inputs must be")
	a.NoError(b.Add(&iface1_1))
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)
}

func TestAddVariety(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(new2ConsumeSlice1))
	a.NoError(b.Add(new3Consume2))
	a.NoError(b.Add(new3ConsumeSlice2))
	a.NoError(b.Add(&iface1_1))
	a.NoError(b.Add(new1Result2))
	a.NoError(b.Add(new1Result3))
	p, errs := b.Build()
	a.Log(errs)
	a.AreEqual(0, len(errs))
	a.NotNil(p)
}

func TestBuildErrorRequiredTypeMissing(t *testing.T) {
	a := assert.New(t)

	b := New()
	a.NoError(b.Add(new3Consume2))
	p, errs := b.Build()
	a.AreEqual(1, len(errs))
	a.ErrorContains(errs[0], "No value of required type")
	a.IsNil(p)
}

func TestBuildProviderReturnsError(t *testing.T) {
	a := assert.New(t)

	b := New()
	a.NoError(b.Add(new1AndErrorError))
	p, errs := b.Build()
	a.AreEqual(1, len(errs))
	a.ErrorContains(errs[0], "new1AndErrorError testing error")
	a.IsNil(p)
}

func TestBuildProviderReturnsErrorInSlice(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(&iface1_1))
	a.NoError(b.Add(new1AndErrorError))
	p, errs := b.Build()
	a.AreEqual(1, len(errs))
	a.ErrorContains(errs[0], "new1AndErrorError testing error")
	a.IsNil(p)
}

func TestBuildProviderReturnsNil(t *testing.T) {
	a := assert.New(t)

	b := New()
	a.NoError(b.Add(new1Nil))
	p, errs := b.Build()
	a.AreEqual(1, len(errs))
	a.ErrorContains(errs[0], "Provider returned nil value")
	a.IsNil(p)
}

func TestBuildErrorCircularRef(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(&iface1_1))
	a.NoError(b.Add(new1Consume2))
	a.NoError(b.Add(new2ConsumeSlice1))
	p, errs := b.Build()
	a.AreEqual(2, len(errs))
	a.ErrorContains(errs[0], "Can't resolve provider")
	a.ErrorContains(errs[1], "Can't resolve provider")
	a.IsNil(p)
}

func TestProviderNotFunc(t *testing.T) {
	a := assert.New(t)

	b := New()
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)

	_, err := p.ProvideFor(1)
	a.ErrorContains(err, "Can only ProvideFor function")
}

func TestProviderSingleValueMadeIntoSlice(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(&iface1_1))
	a.NoError(b.Add(new2ConsumeSlice1))
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)

	sresult, err := p.ProvideFor(toProvideFor)
	a.NoError(err)
	a.AreEqual("testInterface1 #1", sresult)
}

func TestProviderRequiredTypeNotProvided(t *testing.T) {
	a := assert.New(t)

	b := New()
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)

	sresult, err := p.ProvideFor(toProvideFor)
	a.ErrorContains(err, "No value of required type:")
	a.IsNil(sresult)
}

func TestProviderWithVariety(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(new2ConsumeSlice1))
	a.NoError(b.Add(new3Consume2))
	a.NoError(b.Add(new3ConsumeSlice2))
	a.NoError(b.Add(&iface1_1))
	a.NoError(b.Add(new1Result2))
	a.NoError(b.Add(new1Result3))
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)

	sresult, err := p.ProvideFor(toProvideFor)
	a.NoError(err)
	a.AreEqual("testInterface1 #1", sresult)
}

func TestProviderNoReturns(t *testing.T) {
	a := assert.New(t)

	b := New()
	iface1_1 := new1Result1()
	a.NoError(b.Add(&iface1_1))
	p, errs := b.Build()
	a.AreEqual(0, len(errs))
	a.NotNil(p)

	sresult, err := p.ProvideFor(toProvideForNoReturn)
	a.NoError(err)
	a.IsNil(sresult)
}
