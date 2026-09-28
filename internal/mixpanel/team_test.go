package mixpanel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
