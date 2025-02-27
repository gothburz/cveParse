package logging

import (
	"io"
	"log"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle, "[TRACE] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Info = log.New(infoHandle, "[INFO] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Warning = log.New(warningHandle, "[WARNING] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Error = log.New(errorHandle, "[ERROR] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
