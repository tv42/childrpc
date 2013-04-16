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
	log.Println("parent mode")
	client, err := childrpc.RunChild(
		// you could put ssh in here
		"./example",
		[]string{"./example", "--child"},
		os.Environ(),
		".",
		os.Stderr,
	)
	if err != nil {
		log.Fatalf("could not run child: %s", err)
		return
	}
	args := "hello"
	var reply string
	err = client.Call("Echo.Echo", &args, &reply)
	if err != nil {
		log.Fatalf("child call failed: %s", err)
		return
	}
	log.Println("got reply", reply)

	err = client.Close()
	if err != nil {
		log.Fatalf("closing connection to child failed: %s", err)
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
	logger := log.New(os.Stderr, "child: ", log.Ldate|log.Ltime)

	logger.Println("child mode")
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
