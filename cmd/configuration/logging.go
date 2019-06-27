package configuration

import "log"

func SetupLogging() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
