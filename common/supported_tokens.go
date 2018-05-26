package common

import (
	"fmt"
	"strings"
	"sync"
)

type SupportedTokens struct {
	mu          sync.RWMutex
	tokens      []Token          // all active tokens
	idToToken   map[string]Token // map ID to active token
	addrToToken map[string]Token // map address to active token

	iTokens      []Token          // all internal active tokens
	iIDToToken   map[string]Token // map ID to internal active tokens
	iAddrToToken map[string]Token // map address to internal active tokens

	eTokens      []Token          // all external active tokens
	eIDToToken   map[string]Token // map ID to external active tokens
	eAddrToToken map[string]Token // map address to external active tokens

	aTokens      []Token          // all active and inactive tokens
	aIDToToken   map[string]Token // map ID to active or inactive tokens
	aAddrToToken map[string]Token // map address to active or inactive tokens
}

func NewSupportedTokens() *SupportedTokens {
	return &SupportedTokens{
		mu:           sync.RWMutex{},
		tokens:       []Token{},
		idToToken:    map[string]Token{},
		addrToToken:  map[string]Token{},
		iTokens:      []Token{},
		iIDToToken:   map[string]Token{},
		iAddrToToken: map[string]Token{},
		eTokens:      []Token{},
		eIDToToken:   map[string]Token{},
		eAddrToToken: map[string]Token{},
		aTokens:      []Token{},
		aIDToToken:   map[string]Token{},
		aAddrToToken: map[string]Token{},
	}
}

func (self *SupportedTokens) AddInternalActiveToken(t Token) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.tokens = append(self.tokens, t)
	self.idToToken[strings.ToUpper(t.ID)] = t
	self.addrToToken[strings.ToLower(t.Address)] = t

	self.iTokens = append(self.iTokens, t)
	self.iIDToToken[strings.ToUpper(t.ID)] = t
	self.iAddrToToken[strings.ToLower(t.Address)] = t

	self.aTokens = append(self.aTokens, t)
	self.aIDToToken[strings.ToUpper(t.ID)] = t
	self.aAddrToToken[strings.ToLower(t.Address)] = t
}

func (self *SupportedTokens) AddExternalActiveToken(t Token) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.tokens = append(self.tokens, t)
	self.idToToken[strings.ToUpper(t.ID)] = t
	self.addrToToken[strings.ToLower(t.Address)] = t

	self.eTokens = append(self.eTokens, t)
	self.eIDToToken[strings.ToUpper(t.ID)] = t
	self.eAddrToToken[strings.ToLower(t.Address)] = t

	self.aTokens = append(self.aTokens, t)
	self.aIDToToken[strings.ToUpper(t.ID)] = t
	self.aAddrToToken[strings.ToLower(t.Address)] = t
}

func (self *SupportedTokens) AddInactiveToken(t Token) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.aTokens = append(self.aTokens, t)
	self.aIDToToken[strings.ToUpper(t.ID)] = t
	self.aAddrToToken[strings.ToLower(t.Address)] = t
}

func (self *SupportedTokens) GetSupportedTokens() []Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.aTokens
}

func (self *SupportedTokens) GetSupportedTokenByID(id string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.aIDToToken[strings.ToUpper(id)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", id)
	}
	return t, nil
}

func (self *SupportedTokens) GetSupportedTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.aAddrToToken[strings.ToLower(addr)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", addr)
	}
	return t, nil
}

func (self *SupportedTokens) GetTokens() []Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.tokens
}

func (self *SupportedTokens) GetTokenByID(id string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.idToToken[strings.ToUpper(id)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", id)
	}
	return t, nil
}

func (self *SupportedTokens) GetTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.addrToToken[strings.ToLower(addr)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", addr)
	}
	return t, nil
}

func (self *SupportedTokens) GetInternalTokens() []Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.iTokens
}

func (self *SupportedTokens) GetInternalTokenByID(id string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.iIDToToken[strings.ToUpper(id)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", id)
	}
	return t, nil
}

func (self *SupportedTokens) GetInternalTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.iAddrToToken[strings.ToLower(addr)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", addr)
	}
	return t, nil
}

func (self *SupportedTokens) GetExternalTokens() []Token {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.eTokens
}

func (self *SupportedTokens) GetExternalTokenByID(id string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.eIDToToken[strings.ToUpper(id)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", id)
	}
	return t, nil
}

func (self *SupportedTokens) GetExternalTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.eAddrToToken[strings.ToLower(addr)]
	if !found {
		return t, fmt.Errorf("Token %s is not supported by core", addr)
	}
	return t, nil
}

var supportedTokens = NewSupportedTokens()

func ETHToken() Token {
	return MustGetInternalToken("ETH")
}

func RegisterInternalActiveToken(t Token) {
	supportedTokens.AddInternalActiveToken(t)
}

func RegisterExternalActiveToken(t Token) {
	supportedTokens.AddExternalActiveToken(t)
}

func RegisterInactiveToken(t Token) {
	supportedTokens.AddInactiveToken(t)
}

func InternalTokens() []Token {
	return supportedTokens.GetInternalTokens()
}

func NetworkTokens() []Token {
	return supportedTokens.GetTokens()
}

// Get token from SupportedToken and returns error
// if the token is not supported
func GetInternalToken(id string) (Token, error) {
	return supportedTokens.GetInternalTokenByID(id)
}

func GetNetworkToken(id string) (Token, error) {
	return supportedTokens.GetTokenByID(id)
}

func GetNetworkTokenByAddr(addr string) (Token, error) {
	return supportedTokens.GetTokenByAddress(addr)
}

func MustGetInternalToken(id string) Token {
	t, e := GetInternalToken(id)
	if e != nil {
		panic(e)
	}
	return t
}

func GetSupportedTokenByAddr(addr string) (Token, error) {
	return supportedTokens.GetSupportedTokenByAddress(addr)
}

func MustGetSupportedTokenByAddress(addr string) Token {
	t, e := GetSupportedTokenByAddr(addr)
	if e != nil {
		panic(e)
	}
	return t
}
