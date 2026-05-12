package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserSeesOwnPersonalProjectOnly(t *testing.T) {
	store := NewProjectStore()
	own := store.CreatePersonalProject("user-1", "Own", "own")
	store.CreatePersonalProject("user-2", "Other", "other")

	projects := store.AccessibleProjects("user-1")
	if len(projects) != 1 || projects[0].ID != own.ID {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestTeamAndAccountUsersGrantsGiveAccess(t *testing.T) {
	store := NewProjectStore()
	teamProject := store.CreatePersonalProject("owner-1", "Team", "team")
	accountProject := store.CreatePersonalProject("owner-2", "Account", "account")
	store.AddTeamMember("team-1", "user-1")
	store.GrantTeam("team-1", teamProject.ID, RoleRead)
	store.AddAccountMember("account-1", "user-1")
	store.GrantAccountUsers("account-1", accountProject.ID, RoleRead)

	projects := store.AccessibleProjects("user-1")
	if len(projects) != 2 {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestProjectWriteRoleCanSaveModelButReadRoleCannot(t *testing.T) {
	store := NewProjectStore()
	project := store.CreatePersonalProject("owner-1", "Shared", "shared")
	store.GrantUser("reader", project.ID, RoleRead)
	store.GrantUser("writer", project.ID, RoleWrite)

	if err := store.SaveModel("writer", project.ID, "type t"); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveModel("reader", project.ID, "type t"); err == nil {
		t.Fatal("expected read role save to fail")
	}
}

func TestProjectAPIsRequireAuthAndCreateProject(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"New","slug":"new"}`))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:user-1"})
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"slug":"new"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
