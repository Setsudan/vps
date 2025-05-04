package connectionmanager

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Manager holds active WebSocket connections keyed by user ID.
type Manager struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
}

// ConnManager is the global instance for managing connections.
var ConnManager = &Manager{
	connections: make(map[string]*websocket.Conn),
}

// Add registers a new connection for a given userID.
func (m *Manager) Add(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[userID] = conn
}

// Remove deletes the connection associated with a userID.
func (m *Manager) Remove(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, userID)
}

// Get retrieves the connection for a given userID.
func (m *Manager) Get(userID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[userID]
	return conn, ok
}
