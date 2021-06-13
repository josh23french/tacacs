package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

type LDAPConfig struct {
	ServiceUser  string
	ServicePass  string
	GroupMapping map[string][]string
	BaseDN       string
	URL          string
	UseTLS       bool
	InsecureTLS  bool
}

type LDAPBackend struct {
    config      LDAPConfig
    groupLookup map[string]string
}

func New(cfg *LDAPConfig) *LDAPBackend {
    lookup := make(map[string]string)
    for tacacsGroup, ldapGroups := range cfg.GroupMapping {
        for _, ldapGroup := range ldapGroups {
            lookup[ldapGroup] = tacacsGroup
        }
    }
    log.Printf("LDAP Group lookup: %+v\n", lookup)
    return &LDAPBackend{
        config: *cfg,
        groupLookup: lookup,
    }
}

// caller must call conn.Close()!!!
func (be *LDAPBackend) getBoundConn() (*ldap.Conn, error) {
    conn, err := ldap.DialURL(be.config.URL)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if be.config.UseTLS {
		// Reconnect with TLS
		err = conn.StartTLS(&tls.Config{InsecureSkipVerify: be.config.InsecureTLS})
		if err != nil {
			log.Printf("Error during StartTLS: %v\n", err)
			return nil, err
		}
	}

	// First bind with a read only user
	err = conn.Bind(be.config.ServiceUser, be.config.ServicePass)
	if err != nil {
		log.Printf("Error during ServiceUser Bind: %v\n", err)
		return nil, err
	}
	return conn, nil
}

func (be *LDAPBackend) TryAuth(user, pass string) (bool, error) {
    return false, nil
}

func (be *LDAPBackend) GetUserGroups(username string) ([]string, error) {
    conn, err := be.getBoundConn()
    if err != nil {
        return []string(nil), err
    }
    
	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		be.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))", username),
		[]string{"dn", "memberof"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Printf("Error during Search: %v\n", err)
		return []string(nil), err
	}

	if len(sr.Entries) != 1 {
		log.Printf("User does not exist or too many entries returned (%v)", len(sr.Entries))
		return []string(nil), errors.New("User does not exist or too many entries returned")
	}

	user := sr.Entries[0]
	
    userGroups := make([]string, 0)
	
	groups := user.GetEqualFoldAttributeValues("memberOf")
    for _, group := range groups {
        mapping := be.groupLookup[group]
        if mapping != "" {
            userGroups = append(userGroups, mapping)
        }
    }
    return userGroups, nil
}