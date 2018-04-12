package huobi

type Interface interface {
	PublicEndpoint() string
	AuthenticatedEndpoint() string
}

func getOrSetDefaultURL(base_url string) string {
	if len(base_url) > 1 {
		return base_url + ":5200"
	} else {
		return "http://127.0.0.1:5200"
	}

}

type RealInterface struct{}

func (self *RealInterface) PublicEndpoint() string {
	return "https://api.huobi.pro"
}

func (self *RealInterface) AuthenticatedEndpoint() string {
	return "https://api.huobi.pro"
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
	return self.baseurl()
}

func (self *SimulatedInterface) AuthenticatedEndpoint() string {
	return self.baseurl()
}

func NewSimulatedInterface(flagVariable string) *SimulatedInterface {
	return &SimulatedInterface{base_url: flagVariable}
}

type RopstenInterface struct {
	base_url string
}

func (self *RopstenInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *RopstenInterface) PublicEndpoint() string {
	return "https://api.huobi.pro"
}

func (self *RopstenInterface) AuthenticatedEndpoint() string {
	return self.baseurl()
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
	return "https://api.huobi.pro"
}

func (self *KovanInterface) AuthenticatedEndpoint() string {
	return self.baseurl()
}

func NewKovanInterface(flagVariable string) *KovanInterface {
	return &KovanInterface{base_url: flagVariable}
}

type DevInterface struct{}

func (self *DevInterface) PublicEndpoint() string {
	return "https://api.huobi.pro"
	// return "http://192.168.24.247:5200"
	// return "http://192.168.25.16:5100"
}

func (self *DevInterface) AuthenticatedEndpoint() string {
	return "https://api.huobi.pro"
	// return "http://192.168.24.247:5200"
	// return "http://192.168.25.16:5100"
}

func NewDevInterface() *DevInterface {
	return &DevInterface{}
}
