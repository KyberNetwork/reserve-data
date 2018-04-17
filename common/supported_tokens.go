package common

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type SupportedTokens struct {
	mu           sync.RWMutex
	tokens       []Token
	idToToken    map[string]Token
	addrToToken  map[string]Token
	iTokens      []Token
	iIDToToken   map[string]Token
	iAddrToToken map[string]Token
	eTokens      []Token
	eIDToToken   map[string]Token
	eAddrToToken map[string]Token
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
	}
}

func (self *SupportedTokens) AddInternalToken(t Token) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.tokens = append(self.tokens, t)
	self.idToToken[strings.ToUpper(t.ID)] = t
	self.addrToToken[strings.ToLower(t.Address)] = t

	self.iTokens = append(self.tokens, t)
	self.iIDToToken[strings.ToUpper(t.ID)] = t
	self.iAddrToToken[strings.ToLower(t.Address)] = t
}

func (self *SupportedTokens) AddExternalToken(t Token) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.tokens = append(self.tokens, t)
	self.idToToken[strings.ToUpper(t.ID)] = t
	self.addrToToken[strings.ToLower(t.Address)] = t

	self.eTokens = append(self.tokens, t)
	self.eIDToToken[strings.ToUpper(t.ID)] = t
	self.eAddrToToken[strings.ToLower(t.Address)] = t
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
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", id))
	}
	return t, nil
}

func (self *SupportedTokens) GetTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.addrToToken[strings.ToLower(addr)]
	if !found {
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", addr))
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
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", id))
	}
	return t, nil
}

func (self *SupportedTokens) GetInternalTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.iAddrToToken[strings.ToLower(addr)]
	if !found {
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", addr))
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
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", id))
	}
	return t, nil
}

func (self *SupportedTokens) GetExternalTokenByAddress(addr string) (Token, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	t, found := self.eAddrToToken[strings.ToLower(addr)]
	if !found {
		return t, errors.New(fmt.Sprintf("Token %s is not supported by core", addr))
	}
	return t, nil
}

var supportedTokens = NewSupportedTokens()

func ETHToken() Token {
	return MustGetInternalToken("ETH")
}

func RegisterInternalToken(t Token) {
	supportedTokens.AddInternalToken(t)
}

func RegisterExternalToken(t Token) {
	supportedTokens.AddExternalToken(t)
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

func MustGetNetworkTokenByAddress(addr string) Token {
	t, e := GetNetworkTokenByAddr(addr)
	if e != nil {
		panic(e)
	}
	return t
}
