package TeamsClientDeviceLibrary

import (
	"log"
	"os"
)

func SetLogFile(f *os.File) {
	log.SetOutput(f)
}
