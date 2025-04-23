// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

const (
	// DefaultSocketPath is the default Unix socket path
	DefaultSocketPath = "/var/run/snooze.sock"
)

// Request represents a command request sent to the daemon
type Request struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Response represents a response from the daemon
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CommandHandler is a function that handles a command request
type CommandHandler func(params map[string]interface{}) (interface{}, error)

// SocketServer handles the API socket
type SocketServer struct {
	listener   net.Listener
	socketPath string
	handlers   map[string]CommandHandler
	running    bool
}

// SocketClient is a client for communicating with the socket server
type SocketClient struct {
	socketPath string
}

// NewSocketClient creates a new socket client
func NewSocketClient(socketPath string) *SocketClient {
	return &SocketClient{
		socketPath: socketPath,
	}
}

// NewSocketServer creates a new Unix socket server
func NewSocketServer(socketPath string) (*SocketServer, error) {
	// Create socket directory if it doesn't exist
	dir := filepath.Dir(socketPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %v", err)
	}

	// Remove socket file if it already exists
	if err := os.RemoveAll(socketPath); err != nil {
		return nil, fmt.Errorf("failed to remove existing socket: %v", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %v", err)
	}

	// Set permissions on socket file
	if err := os.Chmod(socketPath, 0660); err != nil {
		closeErr := listener.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("failed to set socket permissions: %v, and close listener: %v", err, closeErr)
		}
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
	s.running = true
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.running {
				return nil
			}
			return fmt.Errorf("error accepting connection: %v", err)
		}

		// Handle connection in a goroutine
		go s.handleConnection(conn)
	}
	return nil
}

// Stop stops the socket server
func (s *SocketServer) Stop() error {
	s.running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection processes a client connection
func (s *SocketServer) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

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

	// Execute handler
	result, err := handler(request.Params)
	if err != nil {
		sendErrorResponse(conn, err.Error())
		return
	}

	// Send success response
	response := Response{
		Success: true,
		Data:    result,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		// Not much we can do here since we've already failed to write to the connection
		return
	}
}

// sendErrorResponse sends an error response to the client
func sendErrorResponse(conn net.Conn, errMsg string) {
	response := Response{
		Success: false,
		Error:   errMsg,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		// We're already in an error state, so just log this
		log.Printf("Error sending error response: %v", err)
	}
}

// SendCommand sends a command to the daemon and returns the response
func (c *SocketClient) SendCommand(command string, params map[string]interface{}) (interface{}, error) {
	// Connect to socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}()
	
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
	
	// Check for error
	if !response.Success {
		return nil, fmt.Errorf("daemon error: %s", response.Error)
	}
	
	return response.Data, nil
}