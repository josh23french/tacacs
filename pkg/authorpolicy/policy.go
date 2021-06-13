package authorpolicy

import (
    "log"
    "regexp"
)

type AuthorPolicy struct {
	Group         string   // Group to match
	DenyCommands  []string // Regex of commands to soft-deny (command could be allowed by other policies)
	AllowCommands []string // Regex of commands to allow
}

func (p *AuthorPolicy) Check(command string) bool {
    for _, reg := range p.DenyCommands {
        log.Printf("Checking DenyCommand: %v", reg)
        matched, err := regexp.Match(reg, []byte(command))
        if err != nil || matched {
            log.Printf("Error matching DenyCommand: %v", err)
            return false
        }
    }
    for _, reg := range p.AllowCommands {
        log.Printf("Checking AllowCommand: %v", reg)
        matched, err := regexp.Match(reg, []byte(command))
        if err != nil {
            log.Printf("Error matching AllowCommand: %v", err)
            continue
        }
        if matched {
            return true
        }
    }
    return false
}