package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ExecuteUnfinishErr = errors.New("Execute Unfinish. ")
)

type Param struct {
	Key  string
	Data interface{}
}

type Result struct {
	Key  string
	Err  error
	Data interface{}
}

type ConcurrentPool struct {
	ctx    context.Context
	size   int
	finish bool

	params      []*Param
	genKeyFunc  func(data interface{}) (key string)
	executeFunc func(data interface{}) (resultData interface{}, err error)
	result      []*Result
	execTime    *time.Duration
}

func NewConcurrentPool(size int) *ConcurrentPool {
	return &ConcurrentPool{size: size}
}

func (cp *ConcurrentPool) SetContext(ctx context.Context) *ConcurrentPool {
	cp.ctx = ctx
	return cp
}

func (cp *ConcurrentPool) SetGenKeyFunc(f func(data interface{}) string) *ConcurrentPool {
	cp.genKeyFunc = f
	return cp
}

func (cp *ConcurrentPool) SetParam(data []interface{}) *ConcurrentPool {
	var params []*Param
	for i := 0; i < len(data); i++ {
		params = append(params, &Param{Data: data[i]})
	}
	if cp.genKeyFunc != nil {
		for i := 0; i < len(params); i++ {
			params[i].Key = cp.genKeyFunc(params[i].Data)
		}
	}
	cp.params = params
	return cp
}

func (cp *ConcurrentPool) SetExecuteFunc(f func(data interface{}) (interface{}, error)) *ConcurrentPool {
	cp.executeFunc = f
	return cp
}

func (cp *ConcurrentPool) GetResult() ([]*Result, error) {
	if !cp.finish {
		return nil, ExecuteUnfinishErr
	}
	return cp.result, nil
}

func (cp *ConcurrentPool) GetExecuteTime() (*time.Duration, error) {
	if !cp.finish {
		return nil, ExecuteUnfinishErr
	}

	return cp.execTime, nil
}

func (cp *ConcurrentPool) Execute() (output []*Result, err error) {
	if cp.finish {
		return cp.result, nil
	}
	startTime := time.Now()
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Panic, details: %s", r))
			return
		}
		t := time.Now().Sub(startTime)
		cp.execTime = &t
	}()

	paramChan := make(chan *Param, cp.size)

	go func() {
		for i := 0; i < len(cp.params); i++ {
			paramChan <- cp.params[i]
		}
		close(paramChan)
	}()

	resultChan := make(chan *Result, cp.size)
	consumerWg := sync.WaitGroup{}
	consumerWg.Add(1)
	go func() {
		for {
			result, open := <-resultChan
			if !open {
				break
			}
			cp.result = append(cp.result, result)
		}
		consumerWg.Done()
	}()

	execWg := sync.WaitGroup{}
	execWg.Add(cp.size)
	for i := 0; i < cp.size; i++ {
		go func() {
			for {
				param, open := <-paramChan
				if !open {
					break
				}
				res, err := cp.executeFunc(param.Data)
				resultChan <- &Result{Key: param.Key, Data: res, Err: err}
			}
			execWg.Done()
		}()
	}

	execWg.Wait()
	close(resultChan)
	consumerWg.Wait()

	cp.finish = true
	return cp.result, nil
}
