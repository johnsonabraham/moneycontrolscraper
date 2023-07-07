package errhandler

import (
	"github.com/kataras/golog"
)

func Res(err error) {
	if err != nil {
		golog.Error("failed to handle the responds: ", err)
	}
}
