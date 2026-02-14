package netpkg

import (
	"net"
	"testing"
	"time"
)

func TestUDPClientSendTo(t *testing.T) {
	// Create a test server to receive UDP packets
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer serverConn.Close()
	
	// Get actual server address
	actualAddr := serverConn.LocalAddr().String()
	
	// Test client send
	client := NewUDPClient(time.Second)
	
	_, err = client.SendTo([]byte("Hello UDP"), actualAddr)
	if err != nil {
		t.Fatalf("Failed to send UDP packet: %v", err)
	}
	
	// Receive on server side
	buffer := make([]byte, 1024)
	n, _, err := serverConn.ReadFromUDP(buffer)
	if err != nil {
		t.Fatalf("Failed to receive UDP packet: %v", err)
	}
	
	if string(buffer[:n]) != "Hello UDP" {
		t.Errorf("Expected 'Hello UDP', got '%s'", string(buffer[:n]))
	}
}

func TestUDPServerStartStop(t *testing.T) {
	received := make(chan string, 1)
	
	server := NewUDPServer("127.0.0.1:0", func(data []byte, addr *net.UDPAddr) {
		received <- string(data)
	}, time.Second)
	
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer server.Stop()
	
	// Get actual server address
	actualAddr := server.conn.LocalAddr().String()
	
	// Send a test packet
	client := NewUDPClient(time.Second)
	_, err = client.SendTo([]byte("Test Message"), actualAddr)
	if err != nil {
		t.Fatalf("Failed to send test packet: %v", err)
	}
	
	// Wait for message with timeout
	select {
	case msg := <-received:
		if msg != "Test Message" {
			t.Errorf("Expected 'Test Message', got '%s'", msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for UDP message")
	}
}