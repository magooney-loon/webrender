package database

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

// Pool manages a pool of database connections
type Pool struct {
	// db is the underlying sql.DB instance
	db *sql.DB

	// size is the maximum number of connections in the pool
	size int

	// connections tracks active connections
	connections map[*sql.Conn]bool

	// mu protects concurrent access to the pool
	mu sync.RWMutex

	// healthCheckInterval is how often to check connection health
	healthCheckInterval time.Duration

	// healthCheckTimeout is the timeout for health checks
	healthCheckTimeout time.Duration

	// stop signals the health check goroutine to stop
	stop chan struct{}

	// wg is used to wait for all goroutines to stop
	wg sync.WaitGroup
}

// NewPool creates a new connection pool
func NewPool(db *sql.DB, size int) *Pool {
	if size <= 0 {
		size = 5 // Default pool size
	}

	pool := &Pool{
		db:                  db,
		size:                size,
		connections:         make(map[*sql.Conn]bool),
		healthCheckInterval: 30 * time.Second,
		healthCheckTimeout:  5 * time.Second,
		stop:                make(chan struct{}),
	}

	// Start health check routine
	pool.wg.Add(1)
	go pool.healthCheck()

	return pool
}

// Acquire gets a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (*sql.Conn, error) {
	p.mu.RLock()
	currentCount := len(p.connections)
	p.mu.RUnlock()

	if currentCount >= p.size {
		return nil, errors.New("connection pool exhausted")
	}

	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.connections[conn] = true
	p.mu.Unlock()

	return conn, nil
}

// Release returns a connection to the pool
func (p *Pool) Release(conn *sql.Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.connections[conn]; !exists {
		return errors.New("connection not from this pool")
	}

	delete(p.connections, conn)
	return conn.Close()
}

// Close closes all connections in the pool
func (p *Pool) Close() {
	// Signal health check to stop
	close(p.stop)

	// Wait for goroutine to finish
	p.wg.Wait()

	// Close all connections
	p.mu.Lock()
	defer p.mu.Unlock()

	for conn := range p.connections {
		_ = conn.Close()
	}
	p.connections = make(map[*sql.Conn]bool)
}

// healthCheck periodically checks the health of all connections
func (p *Pool) healthCheck() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkAllConnections()
		case <-p.stop:
			return
		}
	}
}

// checkAllConnections checks the health of all pooled connections
func (p *Pool) checkAllConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for conn := range p.connections {
		// Create a context with timeout for the health check
		ctx, cancel := context.WithTimeout(context.Background(), p.healthCheckTimeout)

		// Ping the connection to check health
		err := conn.PingContext(ctx)
		cancel()

		// If ping fails, close and remove the connection
		if err != nil {
			_ = conn.Close()
			delete(p.connections, conn)
		}
	}
}

// Stats returns statistics about the connection pool
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		ActiveConnections: len(p.connections),
		MaxConnections:    p.size,
	}
}

// PoolStats contains statistics about the connection pool
type PoolStats struct {
	ActiveConnections int
	MaxConnections    int
}
