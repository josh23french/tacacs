package authorizer

import (
    "log"
	"github.com/josh23french/tacacs/pkg/args"
	"github.com/josh23french/tacacs/pkg/config"
	"github.com/josh23french/tacacs/pkg/authorpolicy"
	"github.com/josh23french/tacacs/pkg/backends"
)

type Authorizer struct {
	backends *backends.BackendManager
	policies []authorpolicy.AuthorPolicy
}

func New(cfg *config.Config, backends *backends.BackendManager) *Authorizer {
	return &Authorizer{
		backends: backends,
		policies: cfg.Policy.Authorization,
	}
}

func (a *Authorizer) Auth(username string, args *args.AuthorArgs) (bool, error) {
	groups, err := a.backends.GetUserGroups(username)
	if err != nil {
	    return false, nil
	}
	
	log.Printf("User %v groups: %+v", username, groups)
	
	cmdLine := args.AsShellCommand()
	
	for _, group := range groups {
	    for _, policy := range a.policies {
	        if policy.Group != group {
	            continue // Policy doesn't apply
	        }
    	    log.Printf("Checking Policy: %+v\n", policy)
    	    allowed := policy.Check(cmdLine)
    	    if allowed {
    	        return true, nil
    	    }
    	}
	}
	return false, nil
}