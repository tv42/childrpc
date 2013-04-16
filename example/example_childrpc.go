package main

import (
	"childrpc"
	"flag"
	"log"
	"moreio"
	"net/rpc"
	"os"
)

func parent() {
	log.Stderr("parent mode")
	client, err := childrpc.RunChild(
		// you could put ssh in here
		"./example-childrpc",
		[]string{"./example-childrpc", "--child"},
		os.Environ(),
		".",
		os.Stderr,
	)
	if err != nil {
		log.Exitf("could not run child: %s", err)
		return
	}
	args := "hello"
	var reply string
	err = client.Call("Echo.Echo", &args, &reply)
	if err != nil {
		log.Exitf("child call failed: %s", err)
		return
	}
	log.Stderr("got reply", reply)

	err = client.Close()
	if err != nil {
		log.Exitf("closing connection to child failed: %s", err)
		return
	}
}

// TODO what to use as base type?
type Echo bool

func (t *Echo) Echo(arg *string, reply *string) error {
	*reply = *arg
	return nil
}

func child() {
	logger := log.New(os.Stderr, nil, "child: ", log.Ldate|log.Ltime)

	logger.Log("child mode")
	echo := new(Echo)
	rpc.Register(echo)
	m := moreio.NewReadWriteCloser(os.Stdin, os.Stdout)
	rpc.ServeConn(m)
}

func main() {
	var isChild bool
	flag.BoolVar(&isChild, "child", false, "run in child mode (internal)")
	flag.Parse()

	if isChild {
		child()
	} else {
		parent()
	}
}
