package swim

import (
	"context"
	"errors"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockData struct {
	message string
}

func TestNewTaskRunner(t *testing.T) {
	task1 := func() (interface{}, error) {
		return nil, nil
	}
	ctx1 := context.Background()

	runner := NewTaskRunner(task1, ctx1)

	assert.NotNil(t, runner)
}

func TestTaskRunner_Start_Successfully_Response(t *testing.T) {
	task1 := func() (interface{}, error) {
		return mockData{message: "hello"}, nil
	}

	ctx1, _ := context.WithCancel(context.Background())

	runner := NewTaskRunner(task1, ctx1)

	response := runner.Start()

	payload1 := response.payload
	err1 := response.err

	assert.NoError(t, err1)
	assert.Equal(t, reflect.TypeOf(payload1), reflect.TypeOf(mockData{}))

	data1, ok := payload1.(mockData)
	assert.True(t, ok)
	assert.Equal(t, data1.message, "hello")
}

func TestTaskRunner_Start_Failed_With_Error(t *testing.T) {
	errTaskFailed := errors.New("task failed")
	task1 := func() (interface{}, error) {
		return nil, errTaskFailed
	}

	ctx1, _ := context.WithCancel(context.Background())

	runner := NewTaskRunner(task1, ctx1)

	response := runner.Start()

	payload1 := response.payload
	err1 := response.err

	assert.Error(t, err1, errTaskFailed)
	assert.Nil(t, payload1)
}

func TestTaskRunner_Start_Cancellation(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	errTaskFailed := errors.New("task failed")
	task1 := func() (interface{}, error) {
		time.Sleep(time.Second * 3)
		return mockData{message: "hello"}, errTaskFailed
	}

	ctx1, cancel := context.WithCancel(context.Background())

	runner := NewTaskRunner(task1, ctx1)

	go func() {
		resp := runner.Start()
		assert.Nil(t, resp.payload)
		assert.Nil(t, resp.err)

		wg.Done()
	}()

	// cancel the task
	cancel()
	wg.Wait()
}

func TestParseHostPort(t *testing.T) {
	fullAddr := "111.112.113.114:5555"
	host, port, err := ParseHostPort(fullAddr)
	assert.Equal(t, host, net.ParseIP("111.112.113.114"))
	assert.Equal(t, port, uint16(5555))
	assert.NoError(t, err)

	addr2 := "1234:5555"
	host2, port2, err2 := ParseHostPort(addr2)
	assert.Equal(t, host2, net.IP{})
	assert.Equal(t, uint16(0), port2)
	assert.Error(t, err2)

	addr3 := "111.112.113.114:6666666666"
	host3, port3, err3 := ParseHostPort(addr3)
	assert.Equal(t, host3, net.IP{})
	assert.Equal(t, uint16(0), port3)
	assert.Error(t, err3)
}
