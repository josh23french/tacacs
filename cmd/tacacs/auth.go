package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"

	"github.com/go-ldap/ldap"
	"github.com/josh23french/tacacs/pkg/config"
)

type Authenticator struct {
	BindUser    string
	BindPass    string
	Base        string
	URL         string
	UseTLS      bool
	InsecureTLS bool
}

func NewAuthenticator(cfg *config.Config) *Authenticator {
	return &Authenticator{
		BindUser:    cfg.Backends.LDAP.ServiceUser,
		BindPass:    cfg.Backends.LDAP.ServicePass,
		Base:        cfg.Backends.LDAP.BaseDN,
		URL:         cfg.Backends.LDAP.URL,
		UseTLS:      cfg.Backends.LDAP.UseTLS,
		InsecureTLS: cfg.Backends.LDAP.InsecureTLS,
	}
}

func (a *Authenticator) Auth(username string, password string) (bool, error) {
	l, err := ldap.DialURL(a.URL)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer l.Close()

	if a.UseTLS {
		// Reconnect with TLS
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: a.InsecureTLS})
		if err != nil {
			log.Printf("Error during StartTLS: %v\n", err)
			return false, err
		}
	}

	// First bind with a read only user
	err = l.Bind(a.BindUser, a.BindPass)
	if err != nil {
		log.Printf("Error during ServiceUser Bind: %v\n", err)
		return false, err
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		a.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))", username),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Printf("Error during Search: %v\n", err)
		return false, err
	}

	if len(sr.Entries) != 1 {
		log.Printf("User does not exist or too many entries returned (%v)", len(sr.Entries))
		return false, errors.New("User does not exist or too many entries returned")
	}

	userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		log.Printf("Error during User Bind: %v\n", err)
		return false, err
	}
	return true, nil
}
