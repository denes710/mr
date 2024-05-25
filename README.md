# Simple MapReduce
Distributed MapReduce, consisting of two programs, the coordinator and the worker. There is just one coordinator
process, and one or more worker processes executing in parallel.

## Running
First of all, you should build your distributed application to generate shared object. You can find some example
programs under the `examples` folder. Run the following command to generate their shared object:
```bash
make compile-example-sos
```
After that, you can run your coordinator with an input file with the following command:
```bash
make run-coordinator INPUT_TEXT_FILE=input.txt
```
In another windwo, you can run your reduce with an application logic  with the following command:
```bash
make run-worker SO=wc.so
```

# Simple MapReduce
Distributed MapReduce, consisting of two programs: the coordinator and the worker. There is one coordinator process and
one or more worker processes executing in parallel.

## Running
First, build your distributed application to generate the shared object files. You can find example programs in the
`examples` folder. Run the following command to generate their shared objects:
```bash
make compile-example-sos
```

After that, you can run your coordinator with an input file using the following command:
```bash
make run-coordinator INPUT_TEXT_FILE=input.txt
```

In another window, you can run your worker with the application logic using the following command:
```bash
make run-worker SO=wc.so
```

