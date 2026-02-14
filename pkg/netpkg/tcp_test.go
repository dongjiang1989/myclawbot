package netpkg

import (
	"net"
	"testing"
	"time"
)

func TestTCPClientConnect(t *testing.T) {
	// Start a test server
	server := NewTCPServer("127.0.0.1:0", func(conn net.Conn) {
		conn.Write([]byte("Hello"))
		conn.Close()
	}, time.Second)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
	
	// Get the actual port
	addr := server.listener.Addr().String()
	
	// Test client connection
	client := NewTCPClient(time.Second)
	err = client.Connect(addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	
	buffer := make([]byte, 10)
	n, err := client.Receive(buffer)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}
	
	if string(buffer[:n]) != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", string(buffer[:n]))
	}
}

func TestTCPClientNotConnected(t *testing.T) {
	client := NewTCPClient(time.Second)
	
	_, err := client.Send([]byte("test"))
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
	
	_, err = client.Receive(make([]byte, 10))
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}

func TestTCPServerStartStop(t *testing.T) {
	server := NewTCPServer("127.0.0.1:0", func(conn net.Conn) {
		conn.Close()
	}, time.Second)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	// Should be able to stop without error
	err = server.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}