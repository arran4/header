package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/txtar"
)

// splitArgs splits a string by spaces, respecting single and double quotes.
func splitArgs(s string) []string {
	var args []string
	var current []rune
	inSingleQuote := false
	inDoubleQuote := false
	for _, r := range s {
		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}
		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}
		if r == ' ' && !inSingleQuote && !inDoubleQuote {
			if len(current) > 0 {
				args = append(args, string(current))
				current = current[:0]
			}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		args = append(args, string(current))
	}
	return args
}

func runTxtarScript(t *testing.T, dir string, env map[string]string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

    // Build the binary first
	binPath := filepath.Join(t.TempDir(), "header_bin")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/header")
	buildCmd.Dir = "../.." // root dir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".txt") && !strings.HasSuffix(entry.Name(), ".txtar")) {
			continue
		}

		t.Run(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			var commands string
			var ar *txtar.Archive
			tempDir := t.TempDir()

			ar = txtar.Parse(data)
			for _, f := range ar.Files {
				if f.Name == "commands" {
					commands = string(f.Data)
				} else {
                    targetPath := filepath.Join(tempDir, f.Name)
                    err := os.MkdirAll(filepath.Dir(targetPath), 0755)
                    if err != nil {
                        t.Fatalf("failed to create dir for file: %v", err)
                    }
					err = os.WriteFile(targetPath, f.Data, 0644)
					if err != nil {
						t.Fatalf("failed to write file: %v", err)
					}
				}
			}

			if commands == "" {
				if len(ar.Comment) > 0 {
					commands = string(ar.Comment)
				} else {
					commands = string(data)
				}
			}

            for k, v := range env {
                commands = strings.ReplaceAll(commands, "$"+k, v)
            }

			lines := strings.Split(commands, "\n")
			var lastStdout string
			var stdinFile string

			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				args := splitArgs(line)
				if len(args) == 0 {
					continue
				}

                negate := false
                if args[0] == "!" {
                    negate = true
                    args = args[1:]
                }

				switch args[0] {
				case "stdin":
					if len(args) > 1 {
						stdinFile = filepath.Join(tempDir, args[1])
					}
				case "exec":
					if len(args) < 2 {
						continue
					}

					if args[1] == "cat" {
                        // Implement 'cat' internally for cross-platform support
                        if len(args) < 3 {
                            t.Fatalf("line %d: missing argument for cat", i+1)
                        }
                        content, err := os.ReadFile(filepath.Join(tempDir, args[2]))
                        if err != nil {
                            if !negate {
                                t.Fatalf("line %d: cat failed: %v", i+1, err)
                            }
                        } else {
                            if negate {
                                t.Fatalf("line %d: cat succeeded but should have failed", i+1)
                            }
                        }
                        lastStdout = string(content)
                        continue
					}

					var cmd *exec.Cmd
					if args[1] == "header" {
                        cmd = exec.Command(binPath, args[2:]...)
                        cmd.Dir = tempDir
					} else {
						cmd = exec.Command(args[1], args[2:]...)
                        cmd.Dir = tempDir
					}

                    var out bytes.Buffer
					cmd.Stdout = &out
					cmd.Stderr = os.Stderr

                    var f *os.File
					if stdinFile != "" {
						f, err = os.Open(stdinFile)
						if err != nil {
							t.Fatalf("failed to open stdin file: %v", err)
						}
						cmd.Stdin = f
					}

					err = cmd.Run()

                    if f != nil {
                        f.Close()
                        stdinFile = "" // reset
                    }

					if err != nil {
                        if !negate {
                            if exitErr, ok := err.(*exec.ExitError); ok {
                                t.Fatalf("line %d: exec failed: %v, output: %s", i+1, exitErr, out.String())
                            } else {
                                t.Fatalf("line %d: exec error: %v, output: %s", i+1, err, out.String())
                            }
                        }
					} else {
                        if negate {
                            t.Fatalf("line %d: exec succeeded but was expected to fail. output: %s", i+1, out.String())
                        }
                    }
					lastStdout = out.String()

				case "stdout":
					if len(args) > 1 {
						pattern := args[1]
						matched, err := regexp.MatchString(pattern, lastStdout)
						if err != nil {
							t.Fatalf("invalid regex at line %d %s: %v", i+1, pattern, err)
						}
						if negate {
                            if matched {
                                t.Fatalf("line %d: expected stdout to NOT match %q, but it did", i+1, pattern)
                            }
                        } else {
                            if !matched {
                                t.Fatalf("line %d: expected stdout to match %q, but got %q", i+1, pattern, lastStdout)
                            }
                        }
					}
                default:
                    t.Fatalf("line %d: unknown command: %s", i+1, args[0])
				}
			}
		})
	}
}
