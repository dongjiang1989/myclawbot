package netpkg

import (
	"net"
	"time"
)

// TCPClient represents a TCP client
type TCPClient struct {
	conn    net.Conn
	timeout time.Duration
}

// TCPServer represents a TCP server
type TCPServer struct {
	addr     string
	handler  func(conn net.Conn)
	timeout  time.Duration
	listener net.Listener
}

// NewTCPClient creates a new TCP client
func NewTCPClient(timeout time.Duration) *TCPClient {
	return &TCPClient{
		timeout: timeout,
	}
}

// Connect connects to a TCP server
func (c *TCPClient) Connect(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// Send sends data to the server
func (c *TCPClient) Send(data []byte) (int, error) {
	if c.conn == nil {
		return 0, ErrNotConnected
	}
	
	if c.timeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	
	return c.conn.Write(data)
}

// Receive receives data from the server
func (c *TCPClient) Receive(buffer []byte) (int, error) {
	if c.conn == nil {
		return 0, ErrNotConnected
	}
	
	if c.timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	}
	
	return c.conn.Read(buffer)
}

// Close closes the connection
func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NewTCPServer creates a new TCP server
func NewTCPServer(addr string, handler func(conn net.Conn), timeout time.Duration) *TCPServer {
	return &TCPServer{
		addr:    addr,
		handler: handler,
		timeout: timeout,
	}
}

// Start starts the TCP server
func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Server closed or other error
				return
			}
			
			if s.timeout > 0 {
				conn.SetDeadline(time.Now().Add(s.timeout))
			}
			
			go s.handler(conn)
		}
	}()
	
	return nil
}

// Stop stops the TCP server
func (s *TCPServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}