package configs

import (
	"log"
	"os"
)

var RLogger *log.Logger
var Address []string = []string{"localhost:port"}

func Configure() {
	RLogger = log.New(os.Stdout, "LOGGER: ", log.LstdFlags)
}
