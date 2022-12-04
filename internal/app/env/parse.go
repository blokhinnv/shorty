package env

import (
	"os"
	"strconv"
)

func VarToInt(name string) int {
	v, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		panic(err.Error())
	}
	return v
}
