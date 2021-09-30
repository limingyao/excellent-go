package errs

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrs_Error(t *testing.T) {
	e := New(0, "ok")
	t.Log(e)
	t.Log(fmt.Errorf("%w", e))
}

func TestWrap(t *testing.T) {
	e := New(0, "ok")
	t.Log(e)

	ee := Wrap(New(1, "fail"), e)
	t.Log(ee)

	eee := Wrap(New(2, "err"), e, e)
	t.Log(eee)

	e = errors.New("fatal")
	t.Log(Wrap(New(3, "fatal"), e))
}

func TestErrs_Is(t *testing.T) {
	assert := assert.New(t)

	e := New(0, "ok")

	assert.False(errors.Is(errors.New("error"), e))
	assert.True(errors.Is(e, New(0, "ok")))

	var ee Errs
	assert.True(errors.As(e, &ee))
	t.Log(ee)

	e1 := New(1, "fail")
	e2 := Wrap(e, e1)
	assert.True(errors.Is(e2, e))
}
