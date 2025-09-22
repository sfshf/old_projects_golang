package batch

type Logger interface {
	Complete(err error, params ...any)
	// verbose will only be visible on client.
	VerbosePrint(v ...any)
	InfoPrint(v ...any)
	Progress(line int, total int)
	CountDefinition(count int)
}

type emptyLogger struct{}

func (emptyLogger) Complete(err error, params ...any) {}
func (emptyLogger) VerbosePrint(v ...any)             {}
func (emptyLogger) InfoPrint(v ...any)                {}
func (emptyLogger) Progress(line int, total int)      {}
func (emptyLogger) CountDefinition(count int)         {}

// func (l *Logger) Println(v ...any) {
// 	if atomic.LoadInt32(&l.isDiscard) != 0 {
// 		return
// 	}
// 	l.Output(2, fmt.Sprintln(v...))
// }
