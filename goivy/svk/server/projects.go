package server

import (
	"errors"
	"strconv"
	"sync"
)

type ProjectRole string

const (
	RoleRead  ProjectRole = "read"
	RoleWrite ProjectRole = "write"
	RoleAdmin ProjectRole = "admin"
	RoleOwner ProjectRole = "owner"
)

type ProjectRecord struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	OwnerKind string `json:"ownerKind"`
	OwnerID   string `json:"ownerId"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ProjectStore struct {
	mu              sync.Mutex
	next            int
	projects        map[string]ProjectRecord
	userProjects    map[string]map[string]ProjectRole
	teamMembers     map[string]map[string]bool
	teamProjects    map[string]map[string]ProjectRole
	accountMembers  map[string]map[string]bool
	accountProjects map[string]map[string]ProjectRole
	models          map[string]HostedModel
	results         map[string]HostedJobResult
	snapshots       map[string]HostedGraphSnapshot
}

func NewProjectStore() *ProjectStore {
	return &ProjectStore{
		projects:        map[string]ProjectRecord{},
		userProjects:    map[string]map[string]ProjectRole{},
		teamMembers:     map[string]map[string]bool{},
		teamProjects:    map[string]map[string]ProjectRole{},
		accountMembers:  map[string]map[string]bool{},
		accountProjects: map[string]map[string]ProjectRole{},
		models:          map[string]HostedModel{},
		results:         map[string]HostedJobResult{},
		snapshots:       map[string]HostedGraphSnapshot{},
	}
}

type HostedModel struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Filename  string `json:"filename"`
	Text      string `json:"text"`
	Revision  int64  `json:"revision"`
	UpdatedAt string `json:"updatedAt"`
}

type HostedJobResult struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	JobID     string `json:"jobId"`
	Payload   string `json:"payload"`
}

type HostedGraphSnapshot struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Payload   string `json:"payload"`
}

func (s *ProjectStore) CreatePersonalProject(userID, name, slug string) ProjectRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	project := ProjectRecord{
		ID:        "project-" + stringID(s.next),
		AccountID: "account-" + userID,
		OwnerKind: "user",
		OwnerID:   userID,
		Name:      name,
		Slug:      slug,
		CreatedAt: "2026-05-12T00:00:00.000Z",
		UpdatedAt: "2026-05-12T00:00:00.000Z",
	}
	s.projects[project.ID] = project
	s.grantLocked(userID, project.ID, RoleOwner)
	return project
}

func (s *ProjectStore) GrantUser(userID, projectID string, role ProjectRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grantLocked(userID, projectID, role)
}

func (s *ProjectStore) AddTeamMember(teamID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.teamMembers[teamID] == nil {
		s.teamMembers[teamID] = map[string]bool{}
	}
	s.teamMembers[teamID][userID] = true
}

func (s *ProjectStore) GrantTeam(teamID, projectID string, role ProjectRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.teamProjects[teamID] == nil {
		s.teamProjects[teamID] = map[string]ProjectRole{}
	}
	s.teamProjects[teamID][projectID] = role
}

func (s *ProjectStore) AddAccountMember(accountID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accountMembers[accountID] == nil {
		s.accountMembers[accountID] = map[string]bool{}
	}
	s.accountMembers[accountID][userID] = true
}

func (s *ProjectStore) GrantAccountUsers(accountID, projectID string, role ProjectRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accountProjects[accountID] == nil {
		s.accountProjects[accountID] = map[string]ProjectRole{}
	}
	s.accountProjects[accountID][projectID] = role
}

func (s *ProjectStore) AccessibleProjects(userID string) []ProjectRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen := map[string]bool{}
	var out []ProjectRecord
	for projectID := range s.accessibleRolesLocked(userID) {
		if seen[projectID] {
			continue
		}
		seen[projectID] = true
		if project, ok := s.projects[projectID]; ok {
			out = append(out, project)
		}
	}
	return out
}

func (s *ProjectStore) RoleForUser(userID, projectID string) ProjectRole {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accessibleRolesLocked(userID)[projectID]
}

func (s *ProjectStore) SaveModel(userID, projectID, text string) error {
	_, err := s.SaveModelRevision(userID, projectID, "default.ivy", text, 0)
	return err
}

func (s *ProjectStore) SaveModelRevision(userID, projectID, filename, text string, baseRevision int64) (HostedModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	role := s.accessibleRolesLocked(userID)[projectID]
	if !canWrite(role) {
		return HostedModel{}, errors.New("project write access required")
	}
	key := projectID + ":" + filename
	current := s.models[key]
	if baseRevision != 0 && current.Revision != baseRevision {
		return HostedModel{}, errors.New("stale model revision")
	}
	current.ID = "model-" + projectID + "-" + filename
	current.ProjectID = projectID
	current.Filename = filename
	current.Text = text
	current.Revision++
	current.UpdatedAt = "2026-05-12T00:00:00.000Z"
	s.models[key] = current
	return current, nil
}

func (s *ProjectStore) AppendJobResult(userID, projectID, jobID, payload string) (HostedJobResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accessibleRolesLocked(userID)[projectID] == "" {
		return HostedJobResult{}, errors.New("project access required")
	}
	id := "result-" + projectID + "-" + jobID
	if existing, ok := s.results[id]; ok {
		return existing, nil
	}
	result := HostedJobResult{ID: id, ProjectID: projectID, JobID: jobID, Payload: payload}
	s.results[id] = result
	return result, nil
}

func (s *ProjectStore) AppendGraphSnapshot(userID, projectID, snapshotID, payload string) (HostedGraphSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accessibleRolesLocked(userID)[projectID] == "" {
		return HostedGraphSnapshot{}, errors.New("project access required")
	}
	if _, exists := s.snapshots[snapshotID]; exists {
		return HostedGraphSnapshot{}, errors.New("graph snapshots are immutable")
	}
	snapshot := HostedGraphSnapshot{ID: snapshotID, ProjectID: projectID, Payload: payload}
	s.snapshots[snapshotID] = snapshot
	return snapshot, nil
}

func (s *ProjectStore) grantLocked(userID, projectID string, role ProjectRole) {
	if s.userProjects[userID] == nil {
		s.userProjects[userID] = map[string]ProjectRole{}
	}
	s.userProjects[userID][projectID] = role
}

func (s *ProjectStore) accessibleRolesLocked(userID string) map[string]ProjectRole {
	roles := map[string]ProjectRole{}
	for projectID, role := range s.userProjects[userID] {
		roles[projectID] = maxRole(roles[projectID], role)
	}
	for teamID, members := range s.teamMembers {
		if !members[userID] {
			continue
		}
		for projectID, role := range s.teamProjects[teamID] {
			roles[projectID] = maxRole(roles[projectID], role)
		}
	}
	for accountID, members := range s.accountMembers {
		if !members[userID] {
			continue
		}
		for projectID, role := range s.accountProjects[accountID] {
			roles[projectID] = maxRole(roles[projectID], role)
		}
	}
	return roles
}

func canWrite(role ProjectRole) bool {
	return role == RoleWrite || role == RoleAdmin || role == RoleOwner
}

func maxRole(a, b ProjectRole) ProjectRole {
	order := map[ProjectRole]int{RoleRead: 1, RoleWrite: 2, RoleAdmin: 3, RoleOwner: 4}
	if order[b] > order[a] {
		return b
	}
	return a
}

func stringID(n int) string {
	return strconv.Itoa(n)
}
