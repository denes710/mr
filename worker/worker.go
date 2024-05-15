package worker

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"

	mrrpc "github.com/denes710/mr/rpc"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

type WorkerState int

const (
	Map WorkerState = iota
	Reduce
)

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	myState := Map
	for {
		switch myState {
		case Map:
			// mt.Println("Dealing with map.")

			data, result := CallGetMapJob()
			if !result {
				return
			}
			if data.Done {
				myState = Reduce
				continue
			} else {
				if len(data.Filename) != 0 {
					file, err := os.Open(data.Filename)
					if err != nil {
						log.Fatalf("cannot open %v", data.Filename)
					}
					content, err := ioutil.ReadAll(file)
					if err != nil {
						log.Fatalf("cannot read %v", data.Filename)
					}
					file.Close()

					kva := mapf(data.Filename, string(content))

					partitionedKva := make([][]KeyValue, data.ReducerCount)
					for _, v := range kva {
						partitionKey := ihash(v.Key) % int(data.ReducerCount)
						partitionedKva[partitionKey] = append(partitionedKva[partitionKey], v)
					}

					for i := 0; i < int(data.ReducerCount); i++ {
						oname := fmt.Sprintf("mr-%d-%d", data.MapJobNumber, i)
						ofile, _ := os.Create(oname)
						enc := json.NewEncoder(ofile)

						for _, kv := range partitionedKva[i] {
							err := enc.Encode(&kv)
							if err != nil {
								log.Fatalf("cannot write json output!")
							}
						}
					}

					_, result = CallFinishMapJob(data.MapJobNumber)
					if !result {
						return
					}
				}
			}
		case Reduce:
			// fmt.Println("Dealing with reduce.")
			data, result := CallGetReduceJob()
			if !result {
				return
			}
			if len(data.MappedFiles) == 0 {
				// Waiting
			} else {

				intermediate := []KeyValue{}
				for _, filename := range data.MappedFiles {
					file, err := os.Open(filename)
					if err != nil {
						log.Fatalf("cannot open %v", filename)
					}

					dec := json.NewDecoder(file)
					for {
						var kv KeyValue
						if err := dec.Decode(&kv); err != nil {
							break
						}
						intermediate = append(intermediate, kv)
					}
				}

				sort.Sort(ByKey(intermediate))

				oname := fmt.Sprintf("mr-out-%d", data.ReduceJobNumber)
				ofile, _ := os.Create(oname)

				//
				// call Reduce on each distinct key in intermediate[],
				// and print the result to mr-out-0.
				//
				i := 0
				for i < len(intermediate) {
					j := i + 1
					for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
						j++
					}
					values := []string{}
					for k := i; k < j; k++ {
						values = append(values, intermediate[k].Value)
					}
					output := reducef(intermediate[i].Key, values)

					// this is the correct format for each line of Reduce output.
					fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

					i = j
				}

				ofile.Close()

				_, result = CallFinishReduceJob(data.ReduceJobNumber)
				if !result {
					return
				}
			}
		}

		// waits a little bit before the next action
		time.Sleep(time.Second)
	}
}

func CallGetMapJob() (*mrrpc.GetMapJobResponse, bool) {
	// declare an argument structure.
	args := mrrpc.GetMapJobRequest{}

	// declare a reply structure.
	response := mrrpc.GetMapJobResponse{}

	ok := call("Coordinator.GetMapJob", &args, &response)
	return &response, ok
}

func CallFinishMapJob(mapJobNumber int) (*mrrpc.FinishMapJobResponse, bool) {
	// declare an argument structure.
	args := mrrpc.FinishMapJobRequest{}
	args.MapJobNumber = mapJobNumber

	// declare a reply structure.
	response := mrrpc.FinishMapJobResponse{}

	ok := call("Coordinator.FinishMapJob", &args, &response)
	return &response, ok
}

func CallGetReduceJob() (*mrrpc.GetReduceJobResponse, bool) {
	// declare an argument structure.
	args := mrrpc.GetReduceJobRequest{}

	// declare a reply structure.
	response := mrrpc.GetReduceJobResponse{}

	ok := call("Coordinator.GetReduceJob", &args, &response)
	return &response, ok
}

func CallFinishReduceJob(reduceJobNumber int) (*mrrpc.FinishReduceJobResponse, bool) {
	// declare an argument structure.
	args := mrrpc.FinishReduceJobRequest{}
	args.ReduceJobNumber = reduceJobNumber

	// declare a reply structure.
	response := mrrpc.FinishReduceJobResponse{}

	ok := call("Coordinator.FinishReduceJob", &args, &response)
	return &response, ok
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := mrrpc.CoordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
