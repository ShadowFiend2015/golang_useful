package grpcpool

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func TestGRPCPool(t *testing.T) {
	addr := "127.0.0.1:50051"
	fmt.Println("please start grpc server in 127.0.0.1:50010 in 2 secs")
	time.Sleep(time.Second * 2)
	client, closefunc, err := Get(addr)
	if err != nil || closefunc == nil || client.GetState() != connectivity.Ready {
		t.Fatal("can not get conn")
	}

	if *defaultPool.collections[addr].clients[0].count != 1 {
		t.Fatalf("expect 1 got:%v", *defaultPool.collections[addr].clients[0].count)
	}

	client, closefunc1, err := Get(addr)
	if err != nil || closefunc == nil || client.GetState() != connectivity.Ready {
		t.Fatal("can not get conn")
	}

	if *defaultPool.collections[addr].clients[0].count != 1 {
		t.Fatalf("expect 1 got:%v", *defaultPool.collections[addr].clients[0].count)
	}
	closefunc()
	closefunc1()
	if *defaultPool.collections[addr].clients[0].count != 0 {
		t.Fatalf("expect 0 got:%v", *defaultPool.collections[addr].clients[0].count)
	}

}

func TestServerCount(t *testing.T) {
	addr := "127.0.0.1:50051"
	fmt.Println("please start grpc server in 127.0.0.1:50010 in 2 secs")
	time.Sleep(time.Second * 2)
	client, closefunc, err := Get(addr)
	if err != nil || closefunc == nil || client.GetState() != connectivity.Ready {
		t.Fatal("can not get conn")
	}
	closefunc()

	closefuncs := make([]func(), 10)
	for i := 0; i < 10; i++ {
		_, closefuncs[i], _ = Get(addr)
	}

	for _, v := range defaultPool.collections[addr].clients {
		if *v.count != 2 {
			t.Fatalf("expect 2, got:%v", *v.count)
		}
	}
	for _, v := range closefuncs {
		if v == nil {
			t.Fatal("got nil")
		}
		v()
	}

	for _, v := range defaultPool.collections[addr].clients {
		if *v.count != 0 {
			t.Fatalf("expect 0, got:%v", *v.count)
		}
	}

}

func TestDownServer(t *testing.T) {
	addr := "127.0.0.1:50051"
	fmt.Println("please start grpc server in 127.0.0.1:50010 in 2 secs")
	time.Sleep(time.Second * 2)
	client, closefunc, err := Get(addr)
	if err != nil || closefunc == nil || client.GetState() != connectivity.Ready {
		t.Fatal("can not get conn")
	}
	closefunc()

	conns := make([]*grpc.ClientConn, 10)
	closefuncs := make([]func(), 10)
	for i := 0; i < 10; i++ {
		conns[i], closefuncs[i], _ = Get(addr)
	}

	for _, v := range conns {
		v.Close()
	}

	fmt.Println("first close")
	for _, v := range defaultPool.collections[addr].clients {
		fmt.Println(*v.count)
	}

	for i := 0; i < 10; i++ {
		_, cf, _ := Get(addr)
		cf()
	}

	fmt.Println("get after first close")
	for _, v := range defaultPool.collections[addr].clients {
		fmt.Println(*v.count)
	}

	for _, v := range closefuncs {
		if v == nil {
			t.Fatal("got nil")
		}
		v()
	}

	fmt.Println("cancel after close")
	for _, v := range defaultPool.collections[addr].clients {
		fmt.Println(*v.count)
	}

}

func TestGRPCPoolLeak(t *testing.T) {
	for i := 0; i < 2000; i++ {
		_, _, err := Get("127.0.0.1:50051")
		if err != nil {
			t.Fatal(err)
		}
		//Cl()
	}
	fmt.Println("done")
}
