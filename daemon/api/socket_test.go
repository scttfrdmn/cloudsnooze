// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Test NewSocketServer function
func TestNewSocketServer(t *testing.T) {
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "socket-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")

	// Create the server
	server, err := NewSocketServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket server: %v", err)
	}
	defer server.Stop()

	// Check that the socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatalf("Socket file was not created")
	}

	// Check server properties
	if server.socketPath != socketPath {
		t.Errorf("Expected socketPath to be %s, got %s", socketPath, server.socketPath)
	}

	if server.handlers == nil {
		t.Errorf("Expected handlers map to be initialized")
	}
}

// Test RegisterHandler and command handling
func TestRegisterHandler(t *testing.T) {
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "socket-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")

	// Create the server
	server, err := NewSocketServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket server: %v", err)
	}
	defer server.Stop()

	// Register a test handler
	testValue := "test-value"
	server.RegisterHandler("test", func(params map[string]interface{}) (interface{}, error) {
		return testValue, nil
	})

	// Check that the handler was registered
	if handler, exists := server.handlers["test"]; !exists {
		t.Errorf("Expected handler for 'test' command to be registered")
	} else {
		// Call the handler to make sure it returns the expected value
		result, err := handler(nil)
		if err != nil {
			t.Errorf("Unexpected error from handler: %v", err)
		}
		if result != testValue {
			t.Errorf("Expected handler to return %s, got %v", testValue, result)
		}
	}
}

// setupTestServer creates a test server and starts it
func setupTestServer(t *testing.T) (*SocketServer, string, func()) {
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "socket-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	socketPath := filepath.Join(tempDir, "test.sock")

	// Create the server
	server, err := NewSocketServer(socketPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create socket server: %v", err)
	}

	// Register a test handler
	server.RegisterHandler("echo", func(params map[string]interface{}) (interface{}, error) {
		return params, nil
	})

	// Register a handler that returns an error
	server.RegisterHandler("error", func(params map[string]interface{}) (interface{}, error) {
		return nil, errors.New("test error")
	})

	// Use a channel to signal when server is ready
	serverReady := make(chan struct{})

	// Start the server in a goroutine
	go func() {
		// Signal that the server is starting up
		close(serverReady)
		err := server.Start()
		if err != nil {
			// Only report errors if the server was still supposed to be running
			server.mu.RLock()
			running := server.running
			server.mu.RUnlock()
			if running {
				t.Errorf("Server error: %v", err)
			}
		}
	}()

	// Wait for server to be ready
	<-serverReady
	// Give the server a moment to start listening
	time.Sleep(50 * time.Millisecond)

	// Return a cleanup function
	cleanup := func() {
		server.Stop()
		os.RemoveAll(tempDir)
	}

	return server, socketPath, cleanup
}

// Test client and server communication
func TestClientServerCommunication(t *testing.T) {
	_, socketPath, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a client
	client := NewSocketClient(socketPath)

	// Test sending a command
	params := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	result, err := client.SendCommand("echo", params)
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	// Check the result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result)
	}

	if resultMap["key"] != "value" {
		t.Errorf("Expected key=value, got %v", resultMap["key"])
	}

	// JSON numbers are decoded as float64
	if resultMap["num"] != float64(42) {
		t.Errorf("Expected num=42, got %v", resultMap["num"])
	}
}

// Test the error handling for unknown commands
func TestUnknownCommand(t *testing.T) {
	_, socketPath, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a client
	client := NewSocketClient(socketPath)

	// Test sending an unknown command
	_, err := client.SendCommand("unknown", nil)
	if err == nil {
		t.Fatal("Expected error for unknown command, got nil")
	}
	if err.Error() != "daemon error: Unknown command: unknown" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// Test handler that returns an error
func TestHandlerError(t *testing.T) {
	_, socketPath, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a client
	client := NewSocketClient(socketPath)

	// Test sending a command that will cause a handler error
	_, err := client.SendCommand("error", nil)
	if err == nil {
		t.Fatal("Expected error from handler, got nil")
	}
	if err.Error() != "daemon error: test error" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// Test error when socket path doesn't exist
func TestSocketConnectError(t *testing.T) {
	// Create a client with a non-existent socket path
	client := NewSocketClient("/non/existent/socket.sock")

	// Test sending a command
	_, err := client.SendCommand("echo", nil)
	if err == nil {
		t.Fatal("Expected connection error, got nil")
	}
}

// TestSocketPathCreation tests the directory creation for the socket
func TestSocketPathCreation(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "socket-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a deeper path that doesn't exist yet
	nestedDir := filepath.Join(tempDir, "deep", "nested", "path")
	socketPath := filepath.Join(nestedDir, "test.sock")

	// Create the server - this should create the directory path
	server, err := NewSocketServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket server: %v", err)
	}
	defer server.Stop()

	// Check that the directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Fatalf("Nested directory was not created")
	}

	// Check that the socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatalf("Socket file was not created")
	}
}

// Test manual connection handling for more detailed tests
func TestManualConnection(t *testing.T) {
	_, socketPath, cleanup := setupTestServer(t)
	defer cleanup()

	// Connect to the socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Send a malformed request
	_, err = conn.Write([]byte("not json"))
	if err != nil {
		t.Fatalf("Failed to write to socket: %v", err)
	}

	// Read the response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read from socket: %v", err)
	}

	// Decode the response
	var resp Response
	if err := json.Unmarshal(response[:n], &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that we got an error response
	if resp.Success {
		t.Errorf("Expected success=false for malformed request")
	}
	
	if resp.Error == "" {
		t.Errorf("Expected non-empty error message")
	}
	if resp.Error != "Failed to parse request" {
		t.Errorf("Unexpected error message: %s", resp.Error)
	}
}

// Test concurrent connections
func TestConcurrentConnections(t *testing.T) {
	_, socketPath, cleanup := setupTestServer(t)
	defer cleanup()

	// Number of concurrent clients
	numClients := 10
	// Create a wait group to wait for all goroutines
	wg := sync.WaitGroup{}
	wg.Add(numClients)

	// Use atomic counter to keep track of successful responses
	var successCount atomic.Int32

	// Start concurrent clients
	for i := 0; i < numClients; i++ {
		go func(clientNum int) {
			defer wg.Done()

			// Create a client
			client := NewSocketClient(socketPath)

			// Send a command with a unique parameter
			params := map[string]interface{}{
				"client": clientNum,
				"time":   time.Now().UnixNano(),
			}

			result, err := client.SendCommand("echo", params)
			if err != nil {
				t.Errorf("Client %d failed: %v", clientNum, err)
				return
			}

			// Verify the result contains our client number
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("Client %d: Expected map result, got %T", clientNum, result)
				return
			}

			if int(resultMap["client"].(float64)) != clientNum {
				t.Errorf("Client %d: Got response for client %v", clientNum, resultMap["client"])
				return
			}

			// Increment success counter
			successCount.Add(1)
		}(i)
	}

	// Wait for all clients to complete
	wg.Wait()

	// Check that all clients got successful responses
	if successCount.Load() != int32(numClients) {
		t.Errorf("Expected %d successful responses, got %d", numClients, successCount.Load())
	}
}

// Test server stop and cleanup
func TestServerStop(t *testing.T) {
	server, socketPath, cleanup := setupTestServer(t)
	defer cleanup() // This will still run even if the test fails
	
	// Stop the server directly (not using the cleanup function)
	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
	
	// Give the server a moment to shut down
	time.Sleep(50 * time.Millisecond)
	
	// Check that the socket connection is closed
	_, err := net.Dial("unix", socketPath)
	if err == nil {
		t.Fatal("Expected connection to fail after server stop")
		conn, _ := net.Dial("unix", socketPath)
		defer conn.Close()
	}

	// Verify that the server is marked as not running
	server.mu.RLock()
	running := server.running
	server.mu.RUnlock()

	if running {
		t.Error("Server still marked as running after stop")
	}
}

// Test handle connection edge cases using test mocks
type mockConn struct {
	net.Conn
	readErr  bool
	writeErr bool
	closeErr bool
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.readErr {
		return 0, fmt.Errorf("mock read error")
	}
	
	// Write a valid JSON request
	req := `{"command":"echo","params":{"test":true}}` + "\n"
	copy(b, req)
	return len(req), nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	if m.writeErr {
		return 0, fmt.Errorf("mock write error")
	}
	return len(b), nil
}

func (m *mockConn) Close() error {
	if m.closeErr {
		return fmt.Errorf("mock close error")
	}
	return nil
}

func TestHandleConnectionWriteError(t *testing.T) {
	// Create a temporary server just to get a populated server struct
	tempDir, err := os.MkdirTemp("", "socket-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")
	server, err := NewSocketServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket server: %v", err)
	}
	defer server.Stop()

	// Register the echo handler
	server.RegisterHandler("echo", func(params map[string]interface{}) (interface{}, error) {
		return params, nil
	})

	// Create a mock connection that fails on write
	mock := &mockConn{writeErr: true}

	// The handleConnection method shouldn't panic even with a write error
	server.handleConnection(mock)

	// No assertions needed - we're just testing that it doesn't panic
}