package toolexec

import "strings"

const (
	Asm     = "asm"
	Compile = "compile"
	Link    = "link"
)

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
