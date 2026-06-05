package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func BenchmarkTCPServer(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()
	go func() {
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		for {
			buf := make([]byte, 1024)
			if _, err := conn.Read(buf); err != nil {
				return
			}
			if _, err := conn.Write([]byte("pong")); err != nil {
				return
			}
		}
	}()
	cl, err := net.Dial("tcp", ":8090")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = cl.Close() }()
	b.Run("tcp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = cl.Write([]byte("ping")); err != nil {
				b.Fatal(err)
			}
			buf := make([]byte, 1024)
			if n, err := cl.Read(buf); err != nil {
				b.Fatal(err)
			} else {
				if !slices.Equal(buf[:n], []byte("pong")) {
					b.Fatalf("wrong response %q", buf[:n])
				}
			}
		}
	})
}

func BenchmarkStreamingSockets(b *testing.B) {
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			b.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	addr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		b.Fatal(err)
	}
	cl, err := net.Dial("unix", addr.String())
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = cl.Close() }()
	b.Run("unix_streaming_socket", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = cl.Write([]byte("ping")); err != nil {
				b.Fatal(err)
			}
			buf := make([]byte, 1024)
			if n, err := cl.Read(buf); err != nil {
				b.Fatal(err)
			} else {
				if !slices.Equal(buf[:n], []byte("pong")) {
					b.Fatalf("wrong response %q", buf[:n])
				}
			}
		}
	})
}

func BenchmarkDatagramSockets(b *testing.B) {
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			b.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()
	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		b.Fatal(err)
	}
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	cl, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = cl.Close() }()
	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("unix_datagram_socket", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = cl.WriteTo([]byte("ping"), serverAddr); err != nil {
				b.Fatal(err)
			}
			buf := make([]byte, 1024)
			if n, _, err := cl.ReadFrom(buf); err != nil {
				b.Fatal(err)
			} else {
				if !slices.Equal(buf[:n], []byte("pong")) {
					b.Fatalf("wrong response %q", buf[:n])
				}
			}
		}
	})
}
