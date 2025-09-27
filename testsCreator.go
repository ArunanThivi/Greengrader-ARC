package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Hidden bool   `json:"hidden"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <tests.json> <output_prefix> <command> [args...] [--strict]\n", os.Args[0])
		os.Exit(1)
	}

	jsonFile := os.Args[1]
	outputPrefix := os.Args[2]

	// Parse command and optional --strict flag
	cmdArgs := []string{}
	strictMode := false
	for _, arg := range os.Args[3:] {
		if arg == "--strict" {
			strictMode = true
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	// Read JSON
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", jsonFile, err)
		os.Exit(1)
	}

	var tests []TestCase
	if err := json.Unmarshal(data, &tests); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate one binary per test
	for i, t := range tests {
		binaryName := fmt.Sprintf("%s_test%d", outputPrefix, i+1)

		template := `
			package main

			import (
				"fmt"
				"os"
				"os/exec"
				"strings"
			)

			type TestCase struct {
				Input  string
				Output string
				Hidden bool
			}

			var t = TestCase{
				Input:  %q,
				Output: %q,
				Hidden: %t,
			}

			var defaultCmd = %#v
			var strict = %t

			func runProgram(cmd []string, input string) (string, error) {
				c := exec.Command(cmd[0], cmd[1:]...)
				c.Stdin = strings.NewReader(input)
				out, err := c.Output()
				return strings.TrimSpace(string(out)), err
			}

			func main() {
				cmd := defaultCmd
				if len(os.Args) > 1 {
					cmd = os.Args[1:]
				}

				failed := false
				got, err := runProgram(cmd, t.Input)
				if err != nil {
					fmt.Printf("ERROR running test: %%v\n", err)
					os.Exit(1)
				}

				match := false
				if strict {
					match = got == t.Output
				} else {
					match = strings.Contains(got, t.Output)
				}

				if match {
					if t.Hidden {
						fmt.Printf("PASSED: hidden test\n")
					} else {
						fmt.Printf("PASSED: input='%%s' expected='%%s' got='%%s'\n", t.Input, t.Output, got)
					}
				} else {
					failed = true
					if t.Hidden {
						fmt.Printf("FAILED: hidden test\n")
					} else {
						fmt.Printf("FAILED: input='%%s' expected='%%s' got='%%s'\n", t.Input, t.Output, got)
					}
				}

				if failed {
					os.Exit(1)
				}
			}
		`

		runnerSrc := fmt.Sprintf(template, t.Input, t.Output, t.Hidden, cmdArgs, strictMode)
		tmpFile := fmt.Sprintf("runner_tmp_%d.go", i+1)
		if err := ioutil.WriteFile(tmpFile, []byte(runnerSrc), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write temporary runner for test %d: %v\n", i+1, err)
			continue
		}

		outputPath := fmt.Sprintf("tests/%s", binaryName)
		buildCmd := exec.Command("go", "build", "-o", outputPath, tmpFile)
		buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build binary for test %d: %v\n", i+1, err)
			os.Remove(tmpFile)
			continue
		}

		os.Remove(tmpFile)
		fmt.Printf("Successfully built binary: %s (strict=%t)\n", binaryName, strictMode)
	}
}
