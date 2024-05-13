package command

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
)

func colorize(input string, displayhelp ...bool) string {
	help := false
	if len(displayhelp) > 0 {
		help = displayhelp[0]
	}

	// This is a hack to support the nex tag syntax. The `=` is added during
	// the colorization of tag elements
	input = strings.ReplaceAll(input, "-tags=", "-tags ")

	parts := splitStringWithQuotes(input)
	result := make([]string, len(parts))
	dim := color.New(color.FgWhite, color.Faint).SprintFunc()
	// boldDim := color.New(color.FgWhite, color.Faint, color.Bold).SprintFunc()
	italicDim := color.New(color.FgWhite, color.Faint, color.Italic).SprintFunc()
	tip := italicDim
	cyan := color.New(color.FgCyan).SprintFunc()
	blueBrightBold := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	blueDim := color.New(color.FgBlue, color.Faint).SprintFunc()
	magentaDim := color.New(color.FgMagenta, color.Faint).SprintFunc()
	magentaItalic := color.New(color.FgMagenta, color.Italic).SprintFunc()
	magentaBright := color.New(color.FgHiMagenta).SprintFunc()
	greenDim := color.New(color.FgHiGreen, color.Faint, color.Italic).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	yellowDim := color.New(color.FgYellow, color.Faint).SprintFunc()
	yellowDimItalic := color.New(color.FgYellow, color.Faint, color.Italic).SprintFunc()
	yellowItalic := color.New(color.FgYellow, color.Italic).SprintFunc()
	yellowDimBold := color.New(color.FgYellow, color.Faint, color.Bold).SprintFunc()
	// yellowBright := color.New(color.FgHiYellow).SprintFunc()
	yellowBrightBold := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	base := color.New(color.FgHiCyan).SprintFunc()
	important := color.New(color.FgHiGreen).SprintFunc()
	deprecated := color.New(color.FgHiRed, color.Italic, color.Faint).SprintFunc()(" (deprecated)")

	LDFLAG_PATTERN := `^'([^=]+)=([^']+)'$`
	LDFLAG_RE := regexp.MustCompile(LDFLAG_PATTERN)

	active := "none"

	for i, part := range parts {
		value := part
		startsWithQuote := false
		if part[0:1] == "\"" {
			startsWithQuote = true
		}
		if startsWithQuote && len(part) > 1 {
			part = part[1:]
		}

		endsWithQuote := (part[len(part)-1:] == "\"")
		if endsWithQuote && len(part) > 1 {
			part = part[0 : len(part)-1]
		}

		switch strings.ToLower(strings.TrimSpace(part)) {
		case "tinygo":
			result[i] = greenDim(part)
		case "go":
			result[i] = dim(part)
		case "-ldflags":
			active = "ldflags"
			result[i] = yellowItalic(part)
			if help {
				result[i] += tip(" (link arguments to go tool)")
			}
		case "-x":
			if active == "ldflags" {
				if startsWithQuote {
					result[i] = yellowDim("\"") + dim(" \\\n     ") + yellowDimItalic(part)
				} else {
					result[i] = yellowDimItalic(part)
				}
			} else {
				result[i] = yellowDim(value)
				if help {
					result[i] += tip(" (print the commands)")
				}
			}
		case "-gccgoflags":
			active = "gccgoflags"
			result[i] = yellowItalic(part)
			if help {
				result[i] += tip(" (link arguments to gcc-cgo tool)")
			}
		case "-gcflags":
			active = "gcflags"
			result[i] = yellowItalic(part)
			if help {
				result[i] += tip(" (link arguments to gc tool)")
			}
		case "-asmflags":
			active = "asmflags"
			result[i] = yellowItalic(part)
			if help {
				result[i] += tip(" (link arguments to asm tool)")
			}
		case "-a":
			result[i] = dim(part)
			if help {
				result[i] += tip(" (force rebuild packages that are already up to date)")
			}
		case "-n":
			result[i] = dim(part)
			if help {
				result[i] += tip(" (print commands but do not run them/dry run)")
			}
		case "-o":
			active = "output"
			result[i] = blueDim(part)
			if help {
				result[i] += tip("utput")
			}
		case "-race":
			result[i] = dim(part)
			if help {
				result[i] += tip(" (detect race conditions on Linux)")
			}

		case "-v":
			result[i] = important(part)
			if help {
				result[i] += tip(" (print package names as they are compiled)")
			}
			result[i] += dim(" \\\n  ")
			active = "none"
		case "-tags":
			active = "tags"
			result[i] = magentaItalic(part + "=")
			if help {
				result[i] += tip(" (additional build tags)")
			}
			// result[i] += " " + dim(" \\\n    ")
			result[i] += magentaDim("\"") + dim(" \\\n    ")
		// -i is deprecated
		case "-i":
			active = "none"
			result[i] = base(part)
			if help {
				result[i] += deprecated
			}

			result[i] += dim(" \\\n  ")
		case "&&":
			result[i] = dim(" \\\n") + italicDim(part)
			dim = color.New(color.FgHiWhite, color.Italic).SprintFunc()
		case "-c":
			if active != "ldflags" && active != "gccgoflags" && active != "gcflags" && active != "asmflags" {
				result[i] = blueDim(part)
				if help {
					result[i] += tip("urrent working directory")
				}
				active = "cwd"
				break
			}

			fallthrough
		default:
			switch i {
			case 1:
				result[i] = cyan(part) + dim(" \\") + "\n  "
			default:
				if active == "ldflags" {
					if LDFLAG_RE.Match([]byte(part)) {
						matches := LDFLAG_RE.FindStringSubmatch(part)
						if len(matches) == 3 {
							result[i] = yellowDim("'") + yellowBrightBold(matches[1]) + dim("=") + yellow(matches[2]) + yellowDim("'")
						} else {
							p := strings.SplitN(part, "=", 2)
							if len(p) == 2 {
								result[i] = yellowBrightBold(p[0]) + dim("=") + yellow(p[1])
							} else {
								result[i] = yellowDimBold(part)
							}
						}
					} else {
						result[i] = yellowBrightBold(part)
					}

					if startsWithQuote {
						result[i] = yellowDimBold("\"") + dim(" \\") + "\n    " + result[i]
					}

					if help && part[0:1] == "-" {
						switch part {
						case "-c":
							result[i] += tip("ompress DWARF")
						case "-w":
							result[i] += tip("ithdraw DWARF (strip debugging)")
						case "-s":
							result[i] += tip("trip symbols")
						}
					}

					if endsWithQuote {
						result[i] = result[i] + dim(" \\") + yellowDim("\n   \"") + dim(" \\\n  ")
					} else {
						result[i] = result[i] + dim(" \\") + "\n    "
					}
				} else if active == "gccgoflags" {
					if LDFLAG_RE.Match([]byte(part)) {
						matches := LDFLAG_RE.FindStringSubmatch(part)
						if len(matches) == 3 {
							result[i] = yellowDim("'") + yellowBrightBold(matches[1]) + dim("=") + yellow(matches[2]) + yellowDim("'")
						} else {
							p := strings.SplitN(part, "=", 2)
							if len(p) == 2 {
								result[i] = yellowBrightBold(p[0]) + dim("=") + yellow(p[1])
							} else {
								result[i] = yellowDimBold(part)
							}
						}
					} else {
						result[i] = yellowBrightBold(part)
					}

					if startsWithQuote {
						result[i] = yellowDimBold("\"") + dim(" \\") + "\n    " + result[i]
					}

					if help && part[0:1] == "-" {
						switch part {
						case "-c":
							result[i] += tip("ompress DWARF")
						case "-w":
							result[i] += tip("ithdraw DWARF (strip debugging)")
						case "-s":
							result[i] += tip("trip symbols")
						}
					}

					if endsWithQuote {
						result[i] = result[i] + dim(" \\") + yellowDim("\n   \"") + dim(" \\\n  ")
					} else {
						result[i] = result[i] + dim(" \\") + "\n    "
					}
				} else if active == "gcflags" {
					if LDFLAG_RE.Match([]byte(part)) {
						matches := LDFLAG_RE.FindStringSubmatch(part)
						if len(matches) == 3 {
							result[i] = yellowDim("'") + yellowBrightBold(matches[1]) + dim("=") + yellow(matches[2]) + yellowDim("'")
						} else {
							p := strings.SplitN(part, "=", 2)
							if len(p) == 2 {
								result[i] = yellowBrightBold(p[0]) + dim("=") + yellow(p[1])
							} else {
								result[i] = yellowDimBold(part)
							}
						}
					} else {
						result[i] = yellowBrightBold(part)
					}

					if startsWithQuote {
						result[i] = yellowDimBold("\"") + dim(" \\") + "\n    " + result[i]
					}

					if help && part[0:1] == "-" {
						switch part {
						case "-c":
							result[i] += tip("ompress DWARF")
						case "-w":
							result[i] += tip("ithdraw DWARF (strip debugging)")
						case "-s":
							result[i] += tip("trip symbols")
						}
					}

					if endsWithQuote {
						result[i] = result[i] + dim(" \\") + yellowDim("\n   \"") + dim(" \\\n  ")
					} else {
						result[i] = result[i] + dim(" \\") + "\n    "
					}
				} else if active == "asmflags" {
					if LDFLAG_RE.Match([]byte(part)) {
						matches := LDFLAG_RE.FindStringSubmatch(part)
						if len(matches) == 3 {
							result[i] = yellowDim("'") + yellowBrightBold(matches[1]) + dim("=") + yellow(matches[2]) + yellowDim("'")
						} else {
							p := strings.SplitN(part, "=", 2)
							if len(p) == 2 {
								result[i] = yellowBrightBold(p[0]) + dim("=") + yellow(p[1])
							} else {
								result[i] = yellowDimBold(part)
							}
						}
					} else {
						result[i] = yellowBrightBold(part)
					}

					if startsWithQuote {
						result[i] = yellowDimBold("\"") + dim(" \\") + "\n    " + result[i]
					}

					if help && part[0:1] == "-" {
						switch part {
						case "-c":
							result[i] += tip("ompress DWARF")
						case "-w":
							result[i] += tip("ithdraw DWARF (strip debugging)")
						case "-s":
							result[i] += tip("trip symbols")
						}
					}

					if endsWithQuote {
						result[i] = result[i] + dim(" \\") + yellowDim("\n   \"") + dim(" \\\n  ")
					} else {
						result[i] = result[i] + dim(" \\") + "\n    "
					}
				} else if active == "output" {
					result[i] = blueBrightBold(value)
					active = "none"
				} else if active == "tags" {
					tags := strings.Split(part, ",")
					result[i] = "" //magentaBright(part) + dim(" \\\n   ")
					for ti, tag := range tags {
						comma := " "
						if ti < (len(tags) - 1) {
							comma = ","
						}
						result[i] += magentaBright(tag) + magentaDim(comma) + dim(" \\\n")
						if ti < (len(tags) - 1) {
							result[i] += "     "
						}
					}
					result[i] += magentaDim("    \"") + dim(" \\\n")
					// result[i] = magentaBright(part) + dim(" \\\n   ")
					// if !endsWithQuote {
					// 	result[i] += " "
					// } else {
					// 	result[i] += magentaDim("\"") + dim(" \\\n  ")
					// }

					// result[i] = magentaDim("\"") + dim(" \\\n      ")
					// tags := strings.Split(part, " ")
					// for ti, tag := range tags {
					// 	result[i] += magentaBright(strings.TrimSpace(tag))
					// 	result[i] += dim(" \\") + magentaDim("\n   \"") + dim(" \\\n  ")
					// }
				} else if active == "cwd" {
					result[i] = blueBrightBold(part) + dim(" \\\n  ")
					active = "none"
				} else {
					result[i] = dim(value)
				}
			}
		}

		if endsWithQuote && (active == "ldflags" || active == "gccgoflags" || active == "gcflags" || active == "asmflags") {
			active = "none"
		}
	}

	return strings.Join(result, " ")
}

func splitStringWithQuotes(input string) []string {
	// Define a regular expression pattern to match single-quoted strings
	pattern := `('[^']*')`

	// Find all single-quoted strings in the input
	re := regexp.MustCompile(pattern)
	quotesMatches := re.FindAllString(input, -1)

	// Replace single-quoted strings with a placeholder to protect them from splitting
	placeholder := "<placeholder>"
	replacedInput := re.ReplaceAllString(input, placeholder)

	// Split the remaining string by spaces
	splitResult := strings.Fields(replacedInput)

	// Restore the original single-quoted strings
	for i, part := range splitResult {
		wquote := (placeholder + "\"")
		if part == placeholder || part == wquote {
			splitResult[i] = quotesMatches[0] // Restore the original single-quoted string
			quotesMatches = quotesMatches[1:] // Move to the next single-quoted string
			if part == wquote {
				splitResult[i] = splitResult[i] + "\""
			}
		}
	}

	for i, part := range splitResult {
		if part == placeholder || strings.HasPrefix(part, placeholder) || strings.HasSuffix(part, placeholder) {
			if splitResult[i-1] == "-X" {
				splitResult[i] = ""
				splitResult[i-1] = ""

				if strings.HasSuffix(part, "\"") {
					splitResult[i-2] += "\""
				}
			}
		}
	}

	result := []string{}
	for _, line := range splitResult {
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}
