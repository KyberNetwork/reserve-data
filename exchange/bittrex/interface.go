package bittrex

import (
	"fmt"
)

type Interface interface {
	PublicEndpoint() string
	MarketEndpoint() string
	AccountEndpoint() string
}

type RealInterface struct{}

const apiVersion string = "v1.1"

func getOrSetDefaultURL(base_url string) string {
	if len(base_url) > 1 {
		return base_url + ":5300"
	} else {
		return "http://127.0.0.1:5300"
	}
}

func (self *RealInterface) PublicEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *RealInterface) MarketEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/market"
}

func (self *RealInterface) AccountEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/account"
}

func NewRealInterface() *RealInterface {
	return &RealInterface{}
}

type SimulatedInterface struct {
	base_url string
}

func (self *SimulatedInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *SimulatedInterface) PublicEndpoint() string {
	return fmt.Sprintf("%s/api/%s/public", self.baseurl(), apiVersion)
}

func (self *SimulatedInterface) MarketEndpoint() string {
	return fmt.Sprintf("%s/api/%s/market", self.baseurl(), apiVersion)
}

func (self *SimulatedInterface) AccountEndpoint() string {
	return fmt.Sprintf("%s/api/%s/account", self.baseurl(), apiVersion)
}

func NewSimulatedInterface(flagVariable string) *SimulatedInterface {
	return &SimulatedInterface{base_url: flagVariable}
}

type DevInterface struct{}

func (self *DevInterface) PublicEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *DevInterface) MarketEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/market"
}

func (self *DevInterface) AccountEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/account"
}

func NewDevInterface() *DevInterface {
	return &DevInterface{}
}

type RopstenInterface struct {
	base_url string
}

func (self *RopstenInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *RopstenInterface) PublicEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *RopstenInterface) MarketEndpoint() string {
	return fmt.Sprintf("%s/api/%s/market", self.baseurl(), apiVersion)
}

func (self *RopstenInterface) AccountEndpoint() string {
	return fmt.Sprintf("%s/api/%s/account", self.baseurl(), apiVersion)
}

func NewRopstenInterface(flagVariable string) *RopstenInterface {
	return &RopstenInterface{base_url: flagVariable}
}

type KovanInterface struct {
	base_url string
}

func (self *KovanInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *KovanInterface) PublicEndpoint() string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *KovanInterface) MarketEndpoint() string {
	return fmt.Sprintf("%s/api/%s/market", self.baseurl(), apiVersion)
}

func (self *KovanInterface) AccountEndpoint() string {
	return fmt.Sprintf("%s/api/%s/account", self.baseurl(), apiVersion)
}

func NewKovanInterface(flagVariable string) *KovanInterface {
	return &KovanInterface{base_url: flagVariable}
}
