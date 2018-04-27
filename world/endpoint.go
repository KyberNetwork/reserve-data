package world

type Endpoint interface {
	GoldDataEndpoint() string
	BackupGoldDataEndpoint() string
}

type RealEndpoint struct {
}

func (self RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (self RealEndpoint) BackupGoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

type SimulatedEndpoint struct {
}

func (self SimulatedEndpoint) GoldDataEndpoint() string {
	return ""
}

func (self SimulatedEndpoint) BackupGoldDataEndpoint() string {
	return ""
}
