package toolexec

import (
	"fmt"
	"strings"
)

const (
	Asm     = "asm"
	Compile = "compile"
	Link    = "link"
)

func GetToolexecArgs(args []string) ([]string, error) {
	for i, v := range args {
		if strings.HasSuffix(v, Asm) || strings.HasSuffix(v, Compile) || strings.HasSuffix(v, Link) {
			return args[i-1:], nil
		}
	}
	return nil, fmt.Errorf("no tool found")
}

func GetTool(tool string) string {
	parts := strings.Split(tool, "/")
	return parts[len(parts)-1]
}

func GetPackage(args []string) string {
	for i, v := range args {
		if i+1 == len(args) {
			break
		}

		if v == "-p" {
			return args[i+1]
		}
	}

	return "unknown"
}
