package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type ChatDefaults struct {
	Project string `json:"project"`
	Agent   string `json:"agent"`
}

type StateStore struct {
	stateDir string
}

func NewStateStore(stateDir string) (*StateStore, error) {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("store: create state dir: %w", err)
	}
	return &StateStore{stateDir: stateDir}, nil
}

func (s *StateStore) projectsPath() string {
	return filepath.Join(s.stateDir, "projects.json")
}

func (s *StateStore) chatDefaultsPath() string {
	return filepath.Join(s.stateDir, "chat_defaults.json")
}

// SaveProjects saves the project bindings.
func (s *StateStore) SaveProjects(projects map[string]string) error {
	return AtomicWriteJSON(s.projectsPath(), projects)
}

// LoadProjects loads the project bindings.
func (s *StateStore) LoadProjects() (map[string]string, error) {
	var m map[string]string
	if err := ReadJSON(s.projectsPath(), &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]string)
	}
	return m, nil
}

// BindProject binds a project name to a path and persists.
func (s *StateStore) BindProject(name, path string) error {
	projects, err := s.LoadProjects()
	if err != nil {
		return fmt.Errorf("store: bind project: %w", err)
	}
	projects[name] = path
	return s.SaveProjects(projects)
}

// SetChatDefault sets the default project or agent for a chat.
func (s *StateStore) SetChatDefault(chatID string, key string, value string) error {
	defaults, err := s.loadChatDefaults()
	if err != nil {
		return fmt.Errorf("store: set chat default: %w", err)
	}
	entry := defaults[chatID]
	switch key {
	case "project":
		entry.Project = value
	case "agent":
		entry.Agent = value
	}
	defaults[chatID] = entry
	return s.saveChatDefaults(defaults)
}

// GetChatDefaults returns the saved defaults for a chat.
func (s *StateStore) GetChatDefaults(chatID string) (ChatDefaults, error) {
	defaults, err := s.loadChatDefaults()
	if err != nil {
		return ChatDefaults{}, err
	}
	return defaults[chatID], nil
}

func (s *StateStore) loadChatDefaults() (map[string]ChatDefaults, error) {
	var m map[string]ChatDefaults
	if err := ReadJSON(s.chatDefaultsPath(), &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]ChatDefaults)
	}
	return m, nil
}

func (s *StateStore) saveChatDefaults(defaults map[string]ChatDefaults) error {
	return AtomicWriteJSON(s.chatDefaultsPath(), defaults)
}
