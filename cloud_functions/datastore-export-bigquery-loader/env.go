package bigquery_loader

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
