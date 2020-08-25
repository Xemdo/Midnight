// +build PACKET_DEBUG

package logging

import "log"

func Log_Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Log_Debug(v ...interface{}) {
	log.Print(v...)
}
