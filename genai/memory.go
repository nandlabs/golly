package genai

import (
	"errors"

	"oss.nandlabs.io/golly/managers"
)

const (
	ramMemoryType = "ram"
	ramMemoryId   = "ram-memory"
)

var ErrInvalidSession = errors.New("invalid session")

// Memory is the interface that represents a generative AI memory
type Memory interface {
	// Id returns the id of the memory. This is expected to be unique.
	Id() string
	// Type returns the type of the memory
	Type() string
	// Fetch returns the value of the memory
	Fetch(sessionId, query string) ([]Exchange, error)
	//Last returns the last n message session
	Last(sessionId string, n int) ([]Exchange, error)
	// Set sets the msg of the memory
	Add(sessionId string, exchange Exchange) error
	// Erase erases the value of the memory
	Erase(query string) error
}

// MemoryManager is a manager for memories
var MemoryManager managers.ItemManager[Memory] = managers.NewItemManager[Memory]()

// RamMemory is a memory that stores the data in memory
type RamMemory struct {
	data map[string][]Exchange
}

// NewRamMemory creates a new RamMemory
func NewRamMemory() Memory {
	return &RamMemory{
		data: make(map[string][]Exchange),
	}
}

// Id returns the id of the memory
func (r *RamMemory) Id() string {
	return ramMemoryId
}

// Type returns the type of the memory
func (r *RamMemory) Type() string {
	return ramMemoryType
}

// Fetch returns the value of the memory
func (r *RamMemory) Fetch(sessionId, query string) ([]Exchange, error) {
	//TODO implement query
	if exchanges, ok := r.data[sessionId]; ok {
		return exchanges, nil
	} else {
		return nil, ErrInvalidSession
	}
}

// Last returns the last n message session
func (r *RamMemory) Last(sessionId string, n int) ([]Exchange, error) {
	if exchanges, ok := r.data[sessionId]; ok {
		if len(exchanges) < n || n <= 0 {
			return exchanges, nil
		}
		return exchanges[len(exchanges)-n:], nil
	} else {
		return nil, ErrInvalidSession
	}
}

// Set sets the msg of the memory
func (r *RamMemory) Add(sessionId string, exchange Exchange) error {
	if _, ok := r.data[sessionId]; !ok {
		r.data[sessionId] = []Exchange{}
	}
	var exchanges = r.data[sessionId]
	notFound := true
	for i, e := range exchanges {
		if e.Id() == exchange.Id() {
			// replace the exchange
			exchanges[i] = exchange
			notFound = false
			break
		}
	}
	if notFound {
		exchanges = append(exchanges, exchange)
	}
	r.data[sessionId] = exchanges
	return nil
}

// Erase erases the value of the memory
func (r *RamMemory) Erase(sessionId string) error {
	delete(r.data, sessionId)
	return nil
}
