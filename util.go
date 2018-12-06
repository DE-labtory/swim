package swim

import (
	"context"

	"github.com/it-chain/iLogger"
)

type Task func() (interface{}, error)

type TaskResponse struct {
	payload interface{}
	err     error
}

type TaskRunner struct {
	task Task
	ctx  context.Context
}

func NewTaskRunner(task Task, ctx context.Context) *TaskRunner {
	return &TaskRunner{
		task: task,
		ctx:  ctx,
	}
}

func (t *TaskRunner) Start() TaskResponse {
	done := make(chan TaskResponse)
	defer func() {
		close(done)
	}()

	go func() {
		result, err := t.task()
		if err != nil {
			iLogger.Errorf(nil, "[TaskRunner] error occured: [%s]", err.Error())
			done <- TaskResponse{
				payload: nil,
				err:     err,
			}
			return
		}
		done <- TaskResponse{
			payload: result,
			err:     nil,
		}
	}()

	select {
	case resp := <-done:
		return resp
	case <-t.ctx.Done():
		iLogger.Infof(nil, "[TaskRunner] receive cancel signal, quit")
		return TaskResponse{}
	}
}
