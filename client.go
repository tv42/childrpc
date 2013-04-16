package childrpc

import (
	"github.com/tv42/moreio"
	"net/rpc"
	"os"
	"os/exec"
)

type Child struct {
	cmd    *exec.Cmd
	client *rpc.Client
}

// see rpc.Client.Close
func (child *Child) Close() error {
	err := child.client.Close()
	if err != nil {
		return err
	}

	err = child.cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

// see rpc.Client.Call
func (child *Child) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return child.client.Call(serviceMethod, args, reply)
}

// see rpc.Client.Go
func (child *Child) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return child.client.Go(serviceMethod, args, reply, done)
}

func RunChild(argv0 string, argv []string, envv []string, dir string, stderr *os.File) (*Child, error) {

	cmd := exec.Command(argv0)
	cmd.Args = argv
	cmd.Env = envv
	cmd.Dir = dir
	cmd.Stderr = stderr

	parentW, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	parentR, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		parentW.Close()
		parentR.Close()
		return nil, err
	}

	parent := moreio.NewReadWriteCloser(parentR, parentW)
	client := rpc.NewClient(parent)
	c := Child{
		cmd:    cmd,
		client: client,
	}
	return &c, nil
}
