package configs

import (
	"log"
	"os"
)

var RLogger *log.Logger

func Configure() {
	RLogger = log.New(os.Stdout, "LOGGER: ", log.LstdFlags)
}
