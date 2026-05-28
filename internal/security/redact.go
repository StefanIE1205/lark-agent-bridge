package security

import (
	"regexp"
	"strings"
)

var (
	// KEY=VALUE format (env vars, shell exports)
	reEnvVar = regexp.MustCompile(
		`(?i)((?:export\s+)?(?:[a-z_]*` + secretKeys + `[a-z_]*)\s*=\s*)([^\s"']+)`,
	)

	// "key": "value" JSON format
	reJSON = regexp.MustCompile(
		`(?i)("(?:[a-z_]*` + secretKeys + `[a-z_]*)("\s*:\s*"))([^"]+)`,
	)

	// Bearer <token>
	reBearer = regexp.MustCompile(
		`(?i)(bearer\s+)([a-zA-Z0-9\-._~+/]+=*)`,
	)

	// key value pairs like --api-key <value>
	reCLIFlag = regexp.MustCompile(
		`(?i)(--(?:[a-z-]*` + secretKeys + `[a-z-]*)\s+)([^\s]+)`,
	)
)

const secretKeys = `api[_\-]?key|apikey|secret|token|password|passwd|credential|auth`

func Redact(s string) string {
	if s == "" {
		return s
	}

	s = reEnvVar.ReplaceAllString(s, `${1}***`)
	s = reJSON.ReplaceAllString(s, `${1}***`)
	s = reBearer.ReplaceAllString(s, `${1}***`)
	s = reCLIFlag.ReplaceAllString(s, `${1}***`)

	return s
}

func RedactEnv(env []string) []string {
	out := make([]string, len(env))
	for i, e := range env {
		// Split on first =, redact the value
		idx := strings.Index(e, "=")
		if idx < 0 {
			out[i] = e
			continue
		}
		key := e[:idx]
		if isSecretKey(key) {
			out[i] = key + "=***"
		} else {
			out[i] = e
		}
	}
	return out
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, k := range secretKeyList {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

var secretKeyList = []string{
	"api_key", "apikey", "api-key",
	"secret",
	"token",
	"password", "passwd",
	"credential",
}
