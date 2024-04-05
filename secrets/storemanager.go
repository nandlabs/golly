package secrets

import "sync"

type Manager struct {
	stores map[string]Store
	once   sync.Once
}

func GetManager() *Manager {
	return &Manager{
		stores: nil,
		once:   sync.Once{},
	}
}

func (m *Manager) Register(store Store) {
	if m.stores == nil {
		m.once.Do(func() {
			m.stores = make(map[string]Store)
		})
	}
	m.stores[store.Provider()] = store
}

func (m *Manager) Store(name string) (store Store) {
	if m.stores != nil {
		store = m.stores[name]
	}
	return
}
