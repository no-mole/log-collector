package data_store

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/no-mole/neptune/logger"

	pb "github.com/no-mole/log-collector/protos/logger"
)

func NewUdpDataStore(ctx context.Context, ip string, port int, queueSize int) (Datastore, error) {
	sip := net.ParseIP(ip)
	if sip == nil {
		ips, err := net.LookupIP(ip)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("udp data store unable to parse ip or host:%s", ip)
		}
		sip = ips[0]
	}
	sAddr := &net.UDPAddr{
		IP:   sip,
		Port: port,
	}
	conn, err := net.DialUDP("udp", nil, sAddr)
	if err != nil {
		return nil, err
	}
	return &UdpDataStore{
		ctx:  ctx,
		addr: sAddr,
		conn: conn,
		ch:   make(chan *pb.WriteRequest, queueSize),
	}, nil
}

type UdpDataStore struct {
	ctx  context.Context
	addr *net.UDPAddr
	conn *net.UDPConn
	ch   chan *pb.WriteRequest
	once sync.Once
}

func (u *UdpDataStore) Add(req *pb.WriteRequest) {
	u.once.Do(func() {
		go u.flush()
	})
	u.ch <- req
}

func (u *UdpDataStore) flush() {
	for {
		select {
		case <-u.ctx.Done():
			close(u.ch)
		case req := <-u.ch:
			_, err := u.conn.Write(req.Entry)
			if err != nil {
				logger.Error(u.ctx, "udp_data_store_flush", err)
			}
		}
	}
}
