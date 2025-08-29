package configs

import (
	"log"
	"os"
)

var RLogger *log.Logger
var Address []string = []string{"localhost:19092", "localhost:19093", "localhost:19094"}

func Configure() {
	RLogger = log.New(os.Stdout, "LOGGER: ", log.LstdFlags)
}
