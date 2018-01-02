package common

import (
	"errors"
	"fmt"
	"strings"
)

type Token struct {
	ID      string
	Address string
	Decimal int64
}

type TokenPairPrecision struct {
	Amount int
	Price  int
}

type TokenPairAmountLimit struct {
	Min float32
	Max float32
}

type TokenPairPriceLimit struct {
	Min float32
	Max float32
}

func (self Token) MarshalText() (text []byte, err error) {
	// return []byte(fmt.Sprintf(
	// 	"%s-%s", self.ID, self.Address,
	// )), nil
	return []byte(self.ID), nil
}

func (self Token) IsETH() bool {
	return self.ID == "ETH"
}

type TokenPair struct {
	Base        Token
	Quote       Token
	Precision   TokenPairPrecision
	AmountLimit TokenPairAmountLimit
	PriceLimit  TokenPairPriceLimit
}

func (self *TokenPair) PairID() TokenPairID {
	return NewTokenPairID(self.Base.ID, self.Quote.ID)
}

func NewTokenPair(base, quote string) (TokenPair, error) {
	bToken, err1 := GetToken(base)
	qToken, err2 := GetToken(quote)
	tokenPairPrecision := TokenPairPrecision{0, 0}
	tokenPairAmountLimit := TokenPairAmountLimit{0, 0}
	tokenPairPriceLimit := TokenPairPriceLimit{0, 0}
	if err1 != nil || err2 != nil {
		return TokenPair{}, errors.New(fmt.Sprintf("%s or %s is not supported", base, quote))
	} else {
		return TokenPair{bToken, qToken, tokenPairPrecision, tokenPairAmountLimit, tokenPairPriceLimit}, nil
	}
}

func MustCreateTokenPair(base, quote string) TokenPair {
	pair, err := NewTokenPair(base, quote)
	if err != nil {
		panic(err)
	} else {
		return pair
	}
}

var SupportedTokens map[string]Token

func GetToken(id string) (Token, error) {
	t := SupportedTokens[strings.ToUpper(id)]
	if t.ID == "" {
		return t, errors.New(fmt.Sprintf("Token %s is not supported", id))
	} else {
		return t, nil
	}
}

func MustGetToken(id string) Token {
	t, e := GetToken(id)
	if e != nil {
		panic(e)
	}
	return t
}
