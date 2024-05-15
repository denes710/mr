package rpc

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type GetMapJobRequest struct {
}

type GetMapJobResponse struct {
	Filename     string
	MapJobNumber int
	ReducerCount int
	Done         bool
}

type FinishMapJobRequest struct {
	MapJobNumber int
}

type FinishMapJobResponse struct {
}

type GetReduceJobRequest struct {
}

type GetReduceJobResponse struct {
	MappedFiles     []string
	ReduceJobNumber int
}

type FinishReduceJobRequest struct {
	ReduceJobNumber int
}

type FinishReduceJobResponse struct {
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
// FIXME better socket handling in coordinator and worker side
func CoordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
