package data_exporter

import (
	"fmt"
	"os"
)

func LoadEnvVarOrPanic(envVarName string) string {
	val, hasVal := os.LookupEnv(envVarName)
	if !hasVal {
		panic(fmt.Sprintf("Env variabel %s not found", envVarName))
	}
	return val
}
