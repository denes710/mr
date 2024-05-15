package coordinator

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	mrrpc "github.com/denes710/mr/rpc"
)

type JobState int

const (
	NotUsed JobState = iota
	Processing
	Done
)

type Coordinator struct {
	mapStatus    map[int]JobState
	reduceStatus map[int]JobState
	mapFilenames map[int]string
	nReducer     int
	mut          sync.Mutex
}

func (c *Coordinator) GetMapJob(args *mrrpc.GetMapJobRequest, reply *mrrpc.GetMapJobResponse) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	isAllDone := true

	for id, state := range c.mapStatus {
		if state == NotUsed {
			reply.Filename = c.mapFilenames[id]
			reply.ReducerCount = c.nReducer
			reply.MapJobNumber = id
			c.mapStatus[id] = Processing
			go func() {
				time.Sleep(10 * time.Second)

				c.mut.Lock()
				defer c.mut.Unlock()

				if c.mapStatus[id] != Done {
					c.mapStatus[id] = NotUsed
				}
			}()
			return nil
		}

		if c.mapStatus[id] != Done {
			isAllDone = false
		}
	}

	if isAllDone {
		reply.Done = true
	}

	return nil
}

func (c *Coordinator) FinishMapJob(args *mrrpc.FinishMapJobRequest, reply *mrrpc.FinishMapJobResponse) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.mapStatus[args.MapJobNumber] = Done

	return nil
}

func (c *Coordinator) GetReduceJob(args *mrrpc.GetReduceJobRequest, reply *mrrpc.GetReduceJobResponse) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	for id, state := range c.reduceStatus {
		if state == NotUsed {
			for i := 0; i < len(c.mapStatus); i++ {
				reply.MappedFiles = append(reply.MappedFiles, fmt.Sprintf("mr-%d-%d", i, id))
			}
			reply.ReduceJobNumber = id
			c.reduceStatus[id] = Processing

			go func() {
				time.Sleep(10 * time.Second)

				c.mut.Lock()
				defer c.mut.Unlock()

				if c.reduceStatus[id] != Done {
					c.reduceStatus[id] = NotUsed
				}
			}()

			return nil
		}
	}

	return nil
}

func (c *Coordinator) FinishReduceJob(args *mrrpc.FinishReduceJobRequest, reply *mrrpc.FinishReduceJobResponse) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.reduceStatus[args.ReduceJobNumber] = Done

	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := mrrpc.CoordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mut.Lock()
	defer c.mut.Unlock()

	if len(c.reduceStatus) == 0 {
		for _, state := range c.mapStatus {
			if state != Done {
				return false
			}
		}

		c.reduceStatus = make(map[int]JobState)
		for i := 0; i < c.nReducer; i++ {
			c.reduceStatus[i] = NotUsed
		}

		return false
	}

	for _, state := range c.reduceStatus {
		if state != Done {
			return false
		}
	}

	return true
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{nReducer: nReduce}

	i := 0
	c.mapStatus = make(map[int]JobState)
	c.mapFilenames = make(map[int]string)

	for _, filename := range files {
		c.mapStatus[i] = NotUsed
		c.mapFilenames[i] = filename
		i++
	}

	c.server()
	return &c
}
