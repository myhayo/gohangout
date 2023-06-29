package file_input

import (
	"fmt"
	"log"
)

func Log(v ...interface{}) {
	log.Printf("[File Input Plugin]: %s", fmt.Sprint(v...))
}
func Logf(format string, v ...interface{}) {
	log.Printf("[File Input Plugin]: %s", fmt.Sprintf(format, v...))
}
