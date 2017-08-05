package account

import (
	"authorities"
	"time"
	"fmt"
	"token"
	"strings"
)

type Privilege func(param string) (string, error)

type TLSSignPrivilege struct {
	authority  *authorities.TLSAuthority
	ishost     bool
	lifespan   time.Duration
	commonname string
	dnsnames   []string
}

func NewTLSGrantPrivilege(authority authorities.Authority, ishost bool, lifespan time.Duration, commonname string, dnsnames []string) (Privilege, error) {
	if authority == nil || lifespan < time.Second || commonname == "" {
		return nil, fmt.Errorf("Missing parameter to TLS granting privilege.")
	}
	tauth := authority.(*authorities.TLSAuthority)
	if tauth == nil {
		return nil, fmt.Errorf("TLS granting privilege expects a TLS authority.")
	}
	return func(signing_request string) (string, error) {
		return tauth.Sign(signing_request, ishost, lifespan, commonname, dnsnames)
	}, nil
}

func NewSSHGrantPrivilege(authority authorities.Authority, ishost bool, lifespan time.Duration, keyid string, principals []string) (Privilege, error) {
	if authority == nil || lifespan < time.Second || keyid == "" {
		return nil, fmt.Errorf("Missing parameter to SSH granting privilege.")
	}
	tauth := authority.(*authorities.SSHAuthority)
	if tauth == nil {
		return nil, fmt.Errorf("SSH granting privilege expects a SSH authority.")
	}
	return func(signing_request string) (string, error) {
		return tauth.Sign(signing_request, ishost, lifespan, keyid, principals)
	}, nil
}

func stringInList(value string, within []string) bool {
	for _, elem := range within {
		if value == elem {
			return true
		}
	}
	return false
}

func NewBootstrapPrivilege(allowed_principals []string, lifespan time.Duration, registry *token.TokenRegistry) (Privilege, error) {
	if allowed_principals == nil || lifespan < time.Second {
		return nil, fmt.Errorf("Missing parameter to token granting privilege.")
	}
	return func(encoded_principal string) (string, error) {
		principal := string(encoded_principal)
		if !stringInList(principal, allowed_principals) {
			return "", fmt.Errorf("Principal not allowed to be bootstrapped: %s", encoded_principal)
		}
		generated_token := registry.GrantToken(principal, lifespan)
		return generated_token, nil
	}, nil
}

func NewDelegateAuthorityPrivilege(getAccount func(string) (*Account, error), authority authorities.Authority) (Privilege, error) {
	if authority == nil {
		return nil, fmt.Errorf("Missing parameter to authority delegation privilege.")
	}
	return func(request string) (string, error) {
		parts := strings.SplitN(request, "\n", 3)
		if len(parts) != 3 {
			return "", fmt.Errorf("Malformed delegation request (%d parts).", len(parts))
		}
		new_principal := string(parts[0])
		new_api_request := string(parts[1])
		request = parts[2]
		account, err := getAccount(new_principal)
		if err != nil {
			return "", err
		}
		if account.GrantingAuthority != authority {
			return "", fmt.Errorf("Attempt to delegate outside of allowed authority.")
		}
		return account.InvokeAPIOperation(new_api_request, request)
	}, nil
}

func NewConfigurationPrivilege(contents string) (Privilege, error) {
	return func(request string) (string, error) {
		if len(request) != 0 {
			return "", fmt.Errorf("Expected empty request to configuration endpoint.")
		}
		return contents, nil
	}, nil
}