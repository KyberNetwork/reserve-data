package bittrex

import (
	"fmt"
)

type Interface interface {
	PublicEndpoint(timepoint uint64) string
	MarketEndpoint(timepoint uint64) string
	AccountEndpoint(timepoint uint64) string
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

func (self *RealInterface) PublicEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *RealInterface) MarketEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/market"
}

func (self *RealInterface) AccountEndpoint(timepoint uint64) string {
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

func (self *SimulatedInterface) PublicEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/public?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func (self *SimulatedInterface) MarketEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/market?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func (self *SimulatedInterface) AccountEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/account?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func NewSimulatedInterface(flagVariable string) *SimulatedInterface {
	return &SimulatedInterface{base_url: flagVariable}
}

type DevInterface struct{}

func (self *DevInterface) PublicEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *DevInterface) MarketEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/market"
}

func (self *DevInterface) AccountEndpoint(timepoint uint64) string {
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

func (self *RopstenInterface) PublicEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *RopstenInterface) MarketEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/market?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func (self *RopstenInterface) AccountEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/account?timestamp=%d", self.baseurl(), apiVersion, timepoint)
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

func (self *KovanInterface) PublicEndpoint(timepoint uint64) string {
	return "https://bittrex.com/api/" + apiVersion + "/public"
}

func (self *KovanInterface) MarketEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/market?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func (self *KovanInterface) AccountEndpoint(timepoint uint64) string {
	return fmt.Sprintf("%s/api/%s/account?timestamp=%d", self.baseurl(), apiVersion, timepoint)
}

func NewKovanInterface(flagVariable string) *KovanInterface {
	return &KovanInterface{base_url: flagVariable}
}
