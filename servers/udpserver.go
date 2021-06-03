package servers

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

const maxBufferSize = 1024
const serverAddr = "192.168.0.163:5566"

var timeout = 10 * time.Millisecond

type UDPServer struct {
	Context context.Context
	Address string
}

var instance *UDPServer
var once sync.Once

func Instance() *UDPServer {
	once.Do(func() {
		instance = newUDPServer()
	})
	return instance
}

func newUDPServer() *UDPServer {
	ctx, _ := context.WithCancel(context.Background())
	result := &UDPServer{
		Address: serverAddr,
		Context: ctx,
	}
	return result
}

func (this *UDPServer) Start() (err error) {
	pc, err := net.ListenPacket("udp", this.Address)
	if err != nil {
		return
	}

	defer pc.Close()

	this.handlePacketConn(pc)
	fmt.Println("goroutin exited!")
	return
}

func (this *UDPServer) handlePacketConn(pc net.PacketConn) (err error) {
	fmt.Println("Handle Protobuf request")

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("packet-received: bytes=%d from=%s\n", n, addr.String())

			deadline := time.Now().Add(timeout)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			n, err = pc.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				return
			}
			fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
		}
	}()

	select {
	case <-this.Context.Done():
		fmt.Println("cancelled")
		err = this.Context.Err()
	case err = <-doneChan:
	}
	return
}
