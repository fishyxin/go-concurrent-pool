package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFunc(t *testing.T) {
	type addParam struct {
		a int
		b int
	}
	params := []interface{}{
		addParam{a: 1, b: 2},
		addParam{a: 3, b: 4},
		addParam{a: 4, b: 12},
		addParam{a: 33, b: 3},
		addParam{a: 1000, b: 333},
	}

	cp := NewConcurrentPool(3).
		SetParam(params).
		SetExecuteFunc(func(data interface{}) (interface{}, error) {
			param := data.(addParam)
			return param.a + param.b, nil
		})
	cp.Execute()
	res, err := cp.GetResult()
	assert.Nil(t, err)

	for _, r := range res {
		t.Log(r.Data)
	}
}
