package childrpc

import (
	"fmt"
	"os"
	"rpc"
	"io"
	"moreio"
)

type Child struct {
	pid    int
	socket io.ReadWriteCloser
	client *rpc.Client
}

// see rpc.Client.Close
func (child *Child) Close() os.Error {
	err := child.client.Close()
	if err != nil {
		return err
	}

	wait, err := os.Wait(child.pid, 0)
	if err != nil {
		return err
	}

	if wait.WaitStatus.Exited() {
		status := wait.WaitStatus.ExitStatus()
		if status != 0 {
			s := fmt.Sprintf("child failed with exit code %d", status)
			return os.NewError(s)
		}
	} else if wait.WaitStatus.Signaled() {
		signal := wait.WaitStatus.Signal()
		if signal != 0 {
			s := fmt.Sprintf("child exited due to signal %d", signal)
			return os.NewError(s)
		}
	}
	return nil
}

// see rpc.Client.Call
func (child *Child) Call(serviceMethod string, args interface{}, reply interface{}) os.Error {
	return child.client.Call(serviceMethod, args, reply)
}

// see rpc.Client.Go
func (child *Child) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return child.client.Go(serviceMethod, args, reply, done)
}

func RunChild(argv0 string, argv []string, envv []string, dir string, stderr *os.File) (*Child, os.Error) {
	childR, parentW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	parentR, childW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	pid, err := os.ForkExec(argv0, argv, envv, dir,
		[]*os.File{childR, childW, stderr})
	if err != nil {
		childR.Close()
		parentW.Close()
		parentR.Close()
		childW.Close()
		return nil, nil
	}

	err = childW.Close()
	if err != nil {
		// TODO hoping the child will exit at some point?
		childR.Close()
		parentW.Close()
		parentR.Close()
		return nil, nil
	}
	err = childR.Close()
	if err != nil {
		// TODO hoping the child will exit at some point?
		parentW.Close()
		parentR.Close()
		return nil, nil
	}
	parent := moreio.NewReadWriteCloser(parentR, parentW)
	client := rpc.NewClient(parent)
	c := Child{
		pid:    pid,
		socket: parent,
		client: client,
	}
	return &c, nil
}
