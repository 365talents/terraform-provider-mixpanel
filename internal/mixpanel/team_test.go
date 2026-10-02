package mixpanel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCreateTeam(t *testing.T) {
	created := `[{"name":"b","created":true}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/app/me":
			fmt.Fprint(w, `{"results":{"organizations":{"42":{"id":42}}}}`)
		case "POST /organizations/42/add-teams/":
			fmt.Fprintf(w, `{"results":%s}`, created)
		case "GET /organizations/42/teams/":
			fmt.Fprint(w, `{"results":[{"id":1,"name":"a","projects":[]},{"id":2,"name":"b","projects":[{"id":7,"role":"admin"}]}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	user, secret := "u", "s"
	c, err := NewClient(&user, &secret, 1)
	if err != nil {
		t.Fatal(err)
	}
	c.HostURL = server.URL

	team, err := c.CreateTeam("b")
	if err != nil || team.Id != 2 || team.Projects[0].Role != "admin" {
		t.Fatalf("CreateTeam(b) = %+v, %v", team, err)
	}

	created = `[{"name":"b","created":false}]`
	if _, err := c.CreateTeam("b"); err == nil || !strings.Contains(err.Error(), "already exist") {
		t.Fatalf("CreateTeam of an existing team: err = %v", err)
	}
}

func TestTeamsAreCached(t *testing.T) {
	var listCalls atomic.Int32
	var role atomic.Value
	role.Store("admin")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/app/me":
			fmt.Fprint(w, `{"results":{"organizations":{"42":{"id":42}}}}`)
		case "GET /organizations/42/teams/":
			listCalls.Add(1)
			fmt.Fprintf(w, `{"results":[{"id":1,"name":"a","projects":[{"id":7,"role":%q}]}]}`, role.Load())
		case "POST /organizations/42/add-projects-to-teams/":
			role.Store("analyst")
			fmt.Fprint(w, `{"status":"ok"}`)
		case "POST /organizations/42/teams/1/delete-projects/":
			role.Store("consumer")
			fmt.Fprint(w, `{"status":"ok"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	user, secret := "u", "s"
	c, err := NewClient(&user, &secret, 3)
	if err != nil {
		t.Fatal(err)
	}
	c.HostURL = server.URL

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if team, err := c.GetTeam(1); err != nil || team == nil {
				t.Errorf("GetTeam(1) = %v, %v", team, err)
			}
		})
	}
	wg.Wait()
	if n := listCalls.Load(); n != 1 {
		t.Fatalf("teams listed %d times, want 1", n)
	}

	for _, write := range []struct {
		name string
		do   func() error
		want string
	}{
		{"AddProjectToTeam", func() error { return c.AddProjectToTeam(1, 7, "analyst") }, "analyst"},
		{"RemoveProjectFromTeam", func() error { return c.RemoveProjectFromTeam(1, 7) }, "consumer"},
	} {
		if err := write.do(); err != nil {
			t.Fatal(err)
		}
		team, err := c.GetTeam(1)
		if err != nil || team.Projects[0].Role != write.want {
			t.Fatalf("GetTeam(1) after %s = %+v, %v, want role %s", write.name, team, err, write.want)
		}
	}
	if n := listCalls.Load(); n != 3 {
		t.Fatalf("teams listed %d times, want 3", n)
	}
}
