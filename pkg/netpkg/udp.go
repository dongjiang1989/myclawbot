package netpkg

import (
	"net"
	"time"
)

// UDPClient represents a UDP client
type UDPClient struct {
	conn    *net.UDPConn
	timeout time.Duration
}

// UDPServer represents a UDP server
type UDPServer struct {
	addr     string
	handler  func(data []byte, addr *net.UDPAddr)
	timeout  time.Duration
	conn     *net.UDPConn
	stopChan chan struct{}
}

// NewUDPClient creates a new UDP client
func NewUDPClient(timeout time.Duration) *UDPClient {
	return &UDPClient{
		timeout: timeout,
	}
}

// Connect connects to a UDP server (for connection-oriented UDP)
func (c *UDPClient) Connect(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// SendTo sends data to the specified address
func (c *UDPClient) SendTo(data []byte, addr string) (int, error) {
	if c.conn == nil {
		// Create a temporary connection for this send
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return 0, err
		}
		
		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return 0, err
		}
		defer conn.Close()
		
		if c.timeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(c.timeout))
		}
		
		return conn.Write(data)
	}
	
	// Use existing connection
	if c.timeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	
	return c.conn.Write(data)
}

// ReceiveFrom receives data from any address
func (c *UDPClient) ReceiveFrom(buffer []byte) (int, *net.UDPAddr, error) {
	if c.conn == nil {
		return 0, nil, ErrNotConnected
	}
	
	if c.timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	}
	
	return c.conn.ReadFromUDP(buffer)
}

// Close closes the UDP connection
func (c *UDPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NewUDPServer creates a new UDP server
func NewUDPServer(addr string, handler func(data []byte, addr *net.UDPAddr), timeout time.Duration) *UDPServer {
	return &UDPServer{
		addr:     addr,
		handler:  handler,
		timeout:  timeout,
		stopChan: make(chan struct{}),
	}
}

// Start starts the UDP server
func (s *UDPServer) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.conn = conn
	
	go func() {
		buffer := make([]byte, 65536) // Max UDP packet size
		
		for {
			select {
			case <-s.stopChan:
				return
			default:
				if s.timeout > 0 {
					s.conn.SetReadDeadline(time.Now().Add(s.timeout))
				}
				
				n, addr, err := s.conn.ReadFromUDP(buffer)
				if err != nil {
					// Check if it's a timeout or actual error
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Timeout, continue listening
						continue
					}
					// Other error, stop server
					return
				}
				
				// Handle the received data
				go s.handler(buffer[:n], addr)
			}
		}
	}()
	
	return nil
}

// Stop stops the UDP server
func (s *UDPServer) Stop() error {
	close(s.stopChan)
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}