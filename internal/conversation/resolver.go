package conversation

import (
	"fmt"
	"strings"
)

// Resolver resolves project and agent from context.
type Resolver struct {
	projects []string
	agents   []string
}

// NewResolver creates a new resolver.
func NewResolver(projects, agents []string) *Resolver {
	return &Resolver{
		projects: projects,
		agents:   agents,
	}
}

// ResolveResult contains the resolution result.
type ResolveResult struct {
	Project   string
	Agent     string
	NeedAsk   bool
	Question  string
	Options   []string
}

// ResolveProject resolves the project from various sources.
// Priority: message hint > thread memory > chat memory > single project > ask.
func (r *Resolver) ResolveProject(hint, threadProject, chatProject string) (string, bool, string) {
	// 1. Message hint
	if hint != "" {
		for _, p := range r.projects {
			if strings.EqualFold(p, hint) {
				return p, false, ""
			}
		}
	}

	// 2. Thread memory
	if threadProject != "" {
		return threadProject, false, ""
	}

	// 3. Chat memory
	if chatProject != "" {
		return chatProject, false, ""
	}

	// 4. Single project auto-select
	if len(r.projects) == 1 {
		return r.projects[0], false, ""
	}

	// 5. Multiple projects → ask
	if len(r.projects) > 1 {
		var options []string
		for i, p := range r.projects {
			options = append(options, fmt.Sprintf("%d. %s", i+1, p))
		}
		question := fmt.Sprintf("你要我看哪个项目？\n%s\n回复序号或项目名即可。", strings.Join(options, "\n"))
		return "", true, question
	}

	// No projects configured
	return "", false, "还没有绑定项目。请先用 /bind <name> <path> 绑定一个项目。"
}

// ResolveAgent resolves the agent from various sources.
// Priority: message hint > thread memory > chat memory > task type default > config default.
func (r *Resolver) ResolveAgent(hint, threadAgent, chatAgent, taskType, defaultAgent string) string {
	// 1. Message hint
	if hint != "" {
		for _, a := range r.agents {
			if strings.EqualFold(a, hint) {
				return a
			}
		}
	}

	// 2. Thread memory
	if threadAgent != "" {
		return threadAgent
	}

	// 3. Chat memory
	if chatAgent != "" {
		return chatAgent
	}

	// 4. Task type default
	agentByType := map[string]string{
		"implementation": "codex",
		"debugging":      "codex",
		"review":         "claude",
		"architecture":   "claude",
		"exploration":    "claude",
	}
	if agent, ok := agentByType[taskType]; ok {
		return agent
	}

	// 5. Config default
	return defaultAgent
}

// ParseProjectFromReply parses a project selection from user reply.
// Returns project name and whether it was found.
func (r *Resolver) ParseProjectFromReply(text string) (string, bool) {
	text = strings.TrimSpace(text)

	// Try as number
	for i, p := range r.projects {
		if text == fmt.Sprintf("%d", i+1) {
			return p, true
		}
	}

	// Try as name
	for _, p := range r.projects {
		if strings.EqualFold(p, text) {
			return p, true
		}
	}

	return "", false
}
