package mixpanel

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestServiceAccountClient(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.Method + " " + r.URL.Path {
		case "GET /api/app/me":
			fmt.Fprint(w, `{"results":{"organizations":{"42":{"id":42}}}}`)
		case "GET /api/app/organizations/42/service-accounts":
			fmt.Fprint(w, `{"results":[{"id":1,"username":"a"},{"id":2,"username":"b"}]}`)
		case "POST /api/app/organizations/42/service-accounts/remove-from-project":
			gotBody = string(body)
			fmt.Fprint(w, `{"status":"ok"}`)
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

	if sa, err := c.GetServiceAccount(2); err != nil || sa == nil || sa.Username != "b" {
		t.Fatalf("GetServiceAccount(2) = %v, %v", sa, err)
	}
	if sa, err := c.GetServiceAccount(3); err != nil || sa != nil {
		t.Fatalf("GetServiceAccount(3) = %v, %v, want nil", sa, err)
	}
	if err := c.RemoveServiceAccountFromProject(2, 7); err != nil {
		t.Fatal(err)
	}
	if want := `{"projects":[{"id":7,"service_account_ids":[2]}]}` + "\n"; gotBody != want {
		t.Fatalf("remove body = %q, want %q", gotBody, want)
	}
	if _, err := c.GetProjectServiceAccounts(7); !IsNotFound(err) {
		t.Fatalf("GetProjectServiceAccounts error = %v, want 404", err)
	}
}

func TestOrganizationIdIsCached(t *testing.T) {
	var meCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/app/me":
			meCalls.Add(1)
			fmt.Fprint(w, `{"results":{"organizations":{"42":{"id":42}}}}`)
		case "GET /api/app/organizations/42/service-accounts":
			fmt.Fprint(w, `{"results":[]}`)
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
	for range 20 {
		wg.Go(func() {
			if _, err := c.ListServiceAccounts(); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if n := meCalls.Load(); n != 1 {
		t.Fatalf("GET /api/app/me called %d times, want 1", n)
	}
}

func TestProjectDomainIsCached(t *testing.T) {
	var metadataCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /settings/project/7/metadata":
			metadataCalls.Add(1)
			fmt.Fprint(w, `{"results":{"id":7,"domain":"mixpanel.com"}}`)
		case "GET /api/app/projects/7/service-accounts":
			fmt.Fprint(w, `{"results":[{"id":2,"role":"admin"}]}`)
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

	if _, err := c.GetProject(7); err != nil {
		t.Fatal(err)
	}
	if members, err := c.GetProjectServiceAccounts(7); err != nil || len(members) != 1 {
		t.Fatalf("GetProjectServiceAccounts(7) = %v, %v", members, err)
	}
	if n := metadataCalls.Load(); n != 1 {
		t.Fatalf("project metadata fetched %d times, want 1", n)
	}
}
