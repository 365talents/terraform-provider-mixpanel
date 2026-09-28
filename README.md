# Terraform Provider Scaffolding (Terraform Plugin Framework)

_This template repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework). The template repository built on the [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) can be found at [terraform-provider-scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding). See [Which SDK Should I Use?](https://developer.hashicorp.com/terraform/plugin/framework-benefits) in the Terraform documentation for additional information._

This repository is a *template* for a [Terraform](https://www.terraform.io) provider. It is intended as a starting point for creating Terraform providers, containing:

- A resource and a data source (`internal/provider/`),
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

These files contain boilerplate code that you will need to edit to create your own Terraform provider. Tutorials for creating Terraform providers can be found on the [HashiCorp Developer](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework) platform. _Terraform Plugin Framework specific guides are titled accordingly._

Please see the [GitHub template repository documentation](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template) for how to create a new repository from this template on GitHub.

Once you've written your provider, you'll want to [publish it on the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing) so that others can use it.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

## Acceptance tests

Acceptance tests run against the real Mixpanel API. Each test creates what it needs (service accounts, memberships, ...) and deletes it at the end. If a run is interrupted, delete the leftover `tf-acc-*` resources by hand in the Mixpanel organization settings.

Projects are the exception: only owners can delete them, and service accounts can be at most admin, so the provider can't delete a project. The tests use a persistent test project instead, that they import. Destroying a `mixpanel_project` only removes it from the state, so the project is kept. The tests rename it and set the name back at the end.

Acceptance tests only run when `TF_ACC` is set (`make testacc` sets it), and then fail if one of the variables below is missing.

### Service account

Create a dedicated service account in Mixpanel, under **Organization Settings > Service Accounts**, with the **Admin** organization role. Admin is needed to create projects and manage service accounts.

Use a test organization if you have one: the tests create real resources.

### Test project

Create a project for the tests in Mixpanel, for example `tf-acc`, and keep it. Any domain and timezone work. Add the service account above to it with the **Admin** role.

### Environment variables

| Variable | Description |
|---|---|
| `MIXPANEL_SERVICE_ACCOUNT_USERNAME` | Username of the service account above |
| `MIXPANEL_SERVICE_ACCOUNT_SECRET` | Secret of the service account above |
| `MIXPANEL_TEST_PROJECT_ID` | Id of the test project above |

Locally:

```shell
export MIXPANEL_SERVICE_ACCOUNT_USERNAME=...
export MIXPANEL_SERVICE_ACCOUNT_SECRET=...
export MIXPANEL_TEST_PROJECT_ID=...
make testacc
```

In CI, add them as repository secrets (**Settings > Secrets and variables > Actions**) with the same names. `.github/workflows/test.yml` passes them to the tests. Workflows triggered from forks don't get the secrets, so the tests fail there.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
