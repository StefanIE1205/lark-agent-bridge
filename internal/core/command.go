package core

import (
	"strings"
	"unicode/utf8"
)

type Command struct {
	Name string
	Args []string
	Raw  string
}

func ParseCommand(raw string) Command {
	raw = cleanInput(raw)

	if raw == "" {
		return Command{}
	}

	// ! xxx shorthand for /ask xxx
	if strings.HasPrefix(raw, "!") {
		task := strings.TrimSpace(raw[1:])
		if task == "" {
			return Command{}
		}
		return Command{
			Name: "ask",
			Args: []string{task},
			Raw:  raw,
		}
	}

	// Must start with /
	if !strings.HasPrefix(raw, "/") {
		return Command{}
	}

	rest := raw[1:]
	if rest == "" {
		return Command{}
	}

	// Split into command name and args string
	name, argStr := splitFirst(rest, " ")
	name = strings.ToLower(name)

	switch name {
	case "ping", "help", "status", "sessions", "stop":
		return Command{Name: name, Raw: raw}

	case "ask":
		// /ask <task> — entire remaining text is the task
		task := strings.TrimSpace(argStr)
		if task == "" {
			return Command{Name: "ask", Raw: raw}
		}
		return Command{
			Name: "ask",
			Args: []string{task},
			Raw:  raw,
		}

	case "bind":
		// /bind <name> <path> — path may be quoted
		nameStr, pathStr := parseBindArgs(argStr)
		if nameStr == "" {
			return Command{Name: "bind", Raw: raw}
		}
		args := []string{nameStr}
		if pathStr != "" {
			args = append(args, pathStr)
		}
		return Command{Name: "bind", Args: args, Raw: raw}

	case "project", "agent":
		arg := strings.TrimSpace(argStr)
		if arg == "" {
			return Command{Name: name, Raw: raw}
		}
		return Command{Name: name, Args: []string{arg}, Raw: raw}

	case "log":
		arg := strings.TrimSpace(argStr)
		if arg == "" {
			return Command{Name: "log", Raw: raw}
		}
		return Command{Name: "log", Args: []string{arg}, Raw: raw}

	case "approve", "deny":
		reqID := strings.TrimSpace(argStr)
		if reqID == "" {
			return Command{Name: name, Raw: raw}
		}
		return Command{Name: name, Args: []string{reqID}, Raw: raw}

	default:
		return Command{Name: name, Raw: raw}
	}
}

// cleanInput normalizes input text per spec:
// - trim whitespace
// - convert full-width slashes to /
// - remove @bot mentions (simplified: strip <at>...</at> tags and @xxx patterns)
func cleanInput(s string) string {
	s = strings.TrimSpace(s)
	// Convert full-width characters
	s = strings.ReplaceAll(s, "／", "/")
	s = strings.ReplaceAll(s, "！", "!")
	// Normalize spaces
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// splitFirst splits s by the first occurrence of sep, returning left and right.
// If sep is not found, returns s, "".
func splitFirst(s, sep string) (string, string) {
	idx := strings.Index(s, sep)
	if idx < 0 {
		return s, ""
	}
	return s[:idx], s[idx+len(sep):]
}

// parseBindArgs parses arguments for /bind: <name> [path].
// Supports quoted paths: /bind demo "C:\My Repo"
func parseBindArgs(s string) (string, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}

	// Check if the name is quoted
	if strings.HasPrefix(s, "\"") {
		end := strings.Index(s[1:], "\"")
		if end < 0 {
			return "", ""
		}
		name := s[1 : end+1]
		rest := strings.TrimSpace(s[end+2:])
		path := unquote(rest)
		return name, path
	}

	name, rest := splitFirst(s, " ")
	path := unquote(strings.TrimSpace(rest))
	return name, path
}

// unquote removes surrounding double quotes from s.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func truncStr(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
