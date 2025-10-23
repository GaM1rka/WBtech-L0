package config

import (
	"log"
	"os"
)

var RLogger *log.Logger
var Address []string = []string{"kafka1:29092"}

func Configure() {
	RLogger = log.New(os.Stdout, "LOGGER: ", log.LstdFlags)
}
