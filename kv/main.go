package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

//
// Common RPC request/reply definitions
//
type PutArgs struct {
	Key string
	Value string
}
type PutReply struct {
}
type GetArgs struct {
	Key string
}
type GetReply struct {
	Value string
}

//
// Client
// 
func connect() (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		return nil, fmt.Errorf("dialing RPC server: %w", err)
	}
	return client, nil
}
func put(key string, value string) error {
	client, err := connect()
	if err != nil {
		return err
	}
	defer client.Close()

	args := PutArgs{Key: key, Value: value}
	var reply PutReply
	if err := client.Call("KV.Put", &args, &reply); err != nil {
		return fmt.Errorf("put: %w", err)
	}
	return nil
}
func get(key string) (string, error) {
	client, err := connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	args := GetArgs{Key: key}
	var reply GetReply
	if err := client.Call("KV.Get", &args, &reply); err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	return reply.Value, nil
}
//
// Server
//
type KV struct {
	mu sync.Mutex
	data map[string]string
}
func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[args.Key] = args.Value
	return nil
} 
func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	reply.Value = kv.data[args.Key]
	return nil
}
func server() error {
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	kv := &KV{data: make(map[string]string)}
	rpcs := rpc.NewServer()
	if err := rpcs.Register(kv); err != nil {
		return fmt.Errorf("registering RPC services: %w", err)
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("accepting connection: %w", err)
				return
			}
			go rpcs.ServeConn(conn)
		}
	}()

	return nil
}
//
// Client application
//
func main() {
	if err := server(); err != nil {
		log.Fatal(err)
	}

	if err := put("subject", "6.5840"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(`"put("subject", "6.5840")`)

	value, err := get("subject"); 
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`get("subject") -> `, value)
}

