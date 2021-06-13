package backends

import (
    "log"
	"github.com/josh23french/tacacs/pkg/config"
	"github.com/josh23french/tacacs/pkg/backends/ldap"
)

type Backend interface {
    TryAuth(string, string) (bool, error)
    GetUserGroups(string) ([]string, error)
}

type BackendManager struct {
    backends []Backend
}

func NewManager(configs config.BackendsConfig) *BackendManager {
    backends := make([]Backend, 0)
    if configs.LDAP != nil {
        backends = append(backends, ldap.New(configs.LDAP))
    }
    log.Printf("backends: %+v\n", backends)
    return &BackendManager{
        backends: backends,
    }
}

func (m *BackendManager) TryAuth(user, pass string) (bool, error) {
    for _, backend := range m.backends {
        ok, err := backend.TryAuth(user, pass)
        if err != nil {
            log.Printf("Error trying %v backend: %v", backend, err)
            continue // Try another backend
        }
        return ok, nil
    }
    return false, nil // Default is to deny auth if all backends error
}

func (m *BackendManager) GetUserGroups(user string) ([]string, error) {
    allGroups := make([]string, 0)
    for _, backend := range m.backends {
        groups, err := backend.GetUserGroups(user)
        if err != nil {
            log.Printf("Error trying %v backend: %v", backend, err)
            continue // Try another backend
        }
        allGroups = append(allGroups, groups...)
    }
    return allGroups, nil // Will be empty if no backends return groups!
}