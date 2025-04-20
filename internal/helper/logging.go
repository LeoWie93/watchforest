package helper

import (
	"fmt"
	"log/slog"
	"runtime"
)

func HandleError(err error) (b bool) {
	if err != nil {
		pc, filename, line, _ := runtime.Caller(1)

		slog.Error(fmt.Sprintf("%s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err))
		b = true
	}

	return b
}
