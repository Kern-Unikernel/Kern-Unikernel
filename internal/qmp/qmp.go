package qmp

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	conn net.Conn
}

type Message struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type Response struct {
	Return interface{} `json:"return"`
	Error  *QMPError   `json:"error,omitempty"`
}

type QMPError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type MemoryStats struct {
	ActualBalloon uint64 `json:"actual"`
	RSS           uint64 `json:"rss"`
}

type CPUStats struct {
	CPUTime uint64 `json:"cpu_time"`
}

type Metrics struct {
	MemoryMiB float64
	CPUTime   uint64
	Uptime    time.Duration
}

func FindSocket(pid int) (string, error) {
	// Buscar en directorios de kern directo
	pattern := fmt.Sprintf("/tmp/kern-*/ctrl.sock")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			pidFile := filepath.Join(filepath.Dir(match), "machine.pid")
			data, err := os.ReadFile(pidFile)
			if err != nil {
				continue
			}
			var filePid int
			fmt.Sscanf(string(data), "%d", &filePid)
			if filePid == pid {
				return match, nil
			}
		}
	}

	// Buscar en kraftkit
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	runtimeDir := filepath.Join(home, ".local", "share", "kraftkit", "runtime")
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return "", fmt.Errorf("kraftkit runtime directory not found: %w", err)
	}

	for _, entry := range entries {
		pidFile := filepath.Join(runtimeDir, entry.Name(), "machine.pid")
		data, err := os.ReadFile(pidFile)
		if err != nil {
			continue
		}

		var filePid int
		fmt.Sscanf(string(data), "%d", &filePid)
		if filePid == pid {
			sockPath := filepath.Join(runtimeDir, entry.Name(), "ctrl.sock")
			if _, err := os.Stat(sockPath); err == nil {
				return sockPath, nil
			}
		}
	}

	return "", fmt.Errorf("QMP socket not found for PID %d", pid)
}

func Connect(socketPath string) (*Client, error) {
	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error connecting to QMP socket: %w", err)
	}

	client := &Client{conn: conn}

	// Leer el saludo inicial de QEMU
	buf := make([]byte, 4096)
	if _, err := conn.Read(buf); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading QMP greeting: %w", err)
	}

	// Negociar capacidades
	if err := client.send(Message{Execute: "qmp_capabilities"}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error negotiating QMP capabilities: %w", err)
	}

	if _, err := client.receive(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error in QMP handshake: %w", err)
	}

	return client, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) send(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(data)
	return err
}

func (c *Client) receive() (map[string]interface{}, error) {
	buf := make([]byte, 65536)
	n, err := c.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetMemoryStats() (uint64, error) {
	// Intentar query-memory-size-summary primero
	if err := c.send(Message{Execute: "query-memory-size-summary"}); err != nil {
		return 0, err
	}

	result, err := c.receive()
	if err != nil {
		return 0, err
	}

	if ret, ok := result["return"].(map[string]interface{}); ok {
		if base, ok := ret["base-memory"].(float64); ok {
			return uint64(base), nil
		}
	}

	// Fallback a query-balloon
	if err := c.send(Message{Execute: "query-balloon"}); err != nil {
		return 0, err
	}

	result, err = c.receive()
	if err != nil {
		return 0, err
	}

	if ret, ok := result["return"].(map[string]interface{}); ok {
		if actual, ok := ret["actual"].(float64); ok {
			return uint64(actual), nil
		}
	}

	return 0, fmt.Errorf("could not read memory stats")
}

func (c *Client) GetCPUStats() ([]interface{}, error) {
	if err := c.send(Message{Execute: "query-cpus-fast"}); err != nil {
		return nil, err
	}

	result, err := c.receive()
	if err != nil {
		return nil, err
	}

	ret, ok := result["return"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return ret, nil
}

func (c *Client) GetStatus() (string, error) {
	if err := c.send(Message{Execute: "query-status"}); err != nil {
		return "", err
	}

	result, err := c.receive()
	if err != nil {
		return "", err
	}

	ret, ok := result["return"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	status, ok := ret["status"].(string)
	if !ok {
		return "", fmt.Errorf("could not read status")
	}

	return status, nil
}
