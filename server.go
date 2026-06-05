package main

import (
	"context"
	"net"
	"os"
)

func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
		}()
		conn, err := s.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			_ = buf[:n]
			if _, err = conn.Write([]byte("pong")); err != nil {
				return
			}
		}
	}()
	return s.Addr(), nil
}

func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}
	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
			if network == "unixgram" {
				_ = os.Remove(addr)
			}
		}()
		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}
			_ = buf[:n]
			_, err = s.WriteTo([]byte("pong"), clientAddr)
			if err != nil {
				return
			}
		}
	}()
	return s.LocalAddr(), nil
}
