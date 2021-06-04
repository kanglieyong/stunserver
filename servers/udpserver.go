package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"zspace.cn/stunserver/config"
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
	serverConfig, err := config.Load()
	if err != nil {
		return nil
	}
	result := &UDPServer{
		Address: serverConfig.Server1,
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

			// {"peer_id":"12345","local_address":"192.168.0.163:5055","msg_type":"1"}
			type PeerReq struct {
				PeerID    string `json:"peer_id"`
				LocalAddr string `json:"local_addr"`
				MsgType   string `json:"msg_type"`
			}

			type ProbeResp struct {
				PeerID     string `json:"peer_id"`
				LocalAddr  string `json:"local_addr"`
				RemoteAddr string `json:"remote_addr"`
				MsgType    string `json:"msg_type"`
			}

			var req PeerReq
			err = json.Unmarshal(buffer[:n], &req)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var resp ProbeResp
			resp.PeerID = req.PeerID
			resp.LocalAddr = req.LocalAddr
			resp.RemoteAddr = addr.String()
			resp.MsgType = req.MsgType

			wbuf, err := json.Marshal(resp)
			if err != nil {
				fmt.Println(err)
				continue
			}

			deadline := time.Now().Add(timeout)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			n, err = pc.WriteTo(wbuf, addr)
			if err != nil {
				log.Fatal(err)
				continue
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
