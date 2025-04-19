package api

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	// DefaultSocketPath is the default Unix socket path
	DefaultSocketPath = "/var/run/snooze.sock"
)

// SocketServer handles Unix socket communication
type SocketServer struct {
	listener   net.Listener
	socketPath string
	handlers   map[string]CommandHandler
}

// CommandHandler is a function that handles a command
type CommandHandler func(params map[string]interface{}) (interface{}, error)

// Request represents an API request
type Request struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

// Response represents an API response
type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// NewSocketServer creates a new Unix socket server
func NewSocketServer(socketPath string) (*SocketServer, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}

	// Ensure directory exists
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %v", err)
	}

	// Remove socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing socket: %v", err)
		}
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %v", err)
	}

	// Set permissions on socket file
	if err := os.Chmod(socketPath, 0660); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %v", err)
	}

	return &SocketServer{
		listener:   listener,
		socketPath: socketPath,
		handlers:   make(map[string]CommandHandler),
	}, nil
}

// RegisterHandler registers a command handler
func (s *SocketServer) RegisterHandler(command string, handler CommandHandler) {
	s.handlers[command] = handler
}

// Start starts the socket server
func (s *SocketServer) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("accept error: %v", err)
		}

		go s.handleConnection(conn)
	}
}

// Stop stops the socket server
func (s *SocketServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection processes a client connection
func (s *SocketServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create a decoder for the incoming JSON
	decoder := json.NewDecoder(conn)
	var request Request
	if err := decoder.Decode(&request); err != nil {
		sendErrorResponse(conn, "Failed to parse request")
		return
	}

	// Find handler for the command
	handler, exists := s.handlers[request.Command]
	if !exists {
		sendErrorResponse(conn, fmt.Sprintf("Unknown command: %s", request.Command))
		return
	}

	// Execute the handler
	result, err := handler(request.Params)
	if err != nil {
		sendErrorResponse(conn, err.Error())
		return
	}

	// Send success response
	response := Response{
		Status: "success",
		Data:   result,
	}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		fmt.Printf("Failed to send response: %v\n", err)
	}
}

// sendErrorResponse sends an error response to the client
func sendErrorResponse(conn net.Conn, errMsg string) {
	response := Response{
		Status: "error",
		Error:  errMsg,
	}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		fmt.Printf("Failed to send error response: %v\n", err)
	}
}

// SocketClient handles communication with the daemon
type SocketClient struct {
	socketPath string
}

// NewSocketClient creates a new Unix socket client
func NewSocketClient(socketPath string) *SocketClient {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	
	return &SocketClient{
		socketPath: socketPath,
	}
}

// SendCommand sends a command to the daemon and returns the response
func (c *SocketClient) SendCommand(command string, params map[string]interface{}) (interface{}, error) {
	// Connect to socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %v", err)
	}
	defer conn.Close()
	
	// Create request
	request := Request{
		Command: command,
		Params:  params,
	}
	
	// Send request
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(request); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	
	// Read response
	decoder := json.NewDecoder(conn)
	var response Response
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	// Check response status
	if response.Status == "error" {
		return nil, fmt.Errorf("daemon error: %s", response.Error)
	}
	
	return response.Data, nil
}