package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

const maxBufferSize = 1024

var timeout = 10 * time.Millisecond

func main() {
	fmt.Println("stunserver started!")
	ctx, _ := context.WithCancel(context.Background())
	err := server(ctx, "192.168.50.32:5566")
	if err != nil {
		return
	}
}

func client(ctx context.Context, address string, reader io.Reader) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	defer conn.Close()

	doneChan := make(chan error, 1)

	go func() {
		n, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", n)

		buffer := make([]byte, maxBufferSize)

		deadline := time.Now().Add(timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n", nRead, addr.String())
	}()

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}
	return
}

func server(ctx context.Context, address string) (err error) {
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}

	defer pc.Close()

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
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}
	return
}
