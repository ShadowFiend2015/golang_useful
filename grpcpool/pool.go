package grpcpool

import (
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	defaultPool = NewPool(5)
)

type Factory func(string) (*grpc.ClientConn, error)

type Pool struct {
	capacity    int
	collections map[string]*spool
	sync.Mutex
}

type spool struct {
	addr     string
	capacity int
	clients  []*ClientConn
}

type ClientConn struct {
	conn  *grpc.ClientConn
	count *uint64
	l     *sync.Mutex
}

func NewPool(capacity int) *Pool {
	if capacity < 1 {
		capacity = 1
	}
	return &Pool{
		capacity:    capacity,
		collections: make(map[string]*spool),
	}
}

func newSPool(addr string, capacity int) *spool {
	clients := make([]*ClientConn, capacity)
	for i := 0; i < capacity; i++ {
		clients[i] = &ClientConn{
			l:     &sync.Mutex{},
			count: new(uint64),
		}
	}
	return &spool{
		addr:     addr,
		capacity: capacity,
		clients:  clients,
	}
}

func (p *Pool) Get(addr string) (*grpc.ClientConn, func(), error) {
	log.Println("grpc pool get conn for addr:", addr)
	p.Lock()
	sp, ok := p.collections[addr]
	if !ok {
		sp = newSPool(addr, p.capacity)
		p.collections[addr] = sp
	}
	p.Unlock()

	return sp.get(addr)
}

func (s *spool) get(addr string) (*grpc.ClientConn, func(), error) {
	//find min
	min := s.clients[0]
	for i := 1; i < s.capacity; i++ {
		if *s.clients[i].count < *min.count {
			min = s.clients[i]
		}
	}
	min.l.Lock()
	defer min.l.Unlock()
	if min.conn == nil || min.conn.GetState() != connectivity.Ready {
		if min.conn != nil {
			min.conn.Close()
		}
		log.Println("grpc pool new conn for addr:", addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}
		min.conn = conn
		min.count = new(uint64)
	}
	*min.count += 1
	oricount := min.count
	return min.conn, func() {
		*oricount--
	}, nil
}

func Get(addr string) (*grpc.ClientConn, func(), error) {
	return defaultPool.Get(addr)
}
