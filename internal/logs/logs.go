package logs

import (
	"fmt"
	"log"
)

func LogAndPrint(format string, args ...any) {
	log.Printf(format, args...)
	fmt.Printf(format, args...)
}
