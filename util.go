package TeamsClientDeviceLibrary

import (
	"io"
	"log"
)

func SetLogFile(f io.Writer) {
	log.SetOutput(f)
}
