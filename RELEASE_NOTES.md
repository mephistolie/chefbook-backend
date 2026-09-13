# Backend release candidate

This coordinated release introduces structured service logging, `username` in
authentication, and `displayName` in user/profile data. Service and API tags use
`rc.1` until deployment and client integration have been validated.

## Compatibility

- Auth RPCs and HTTP routes use username terminology instead of nickname.
- Access tokens use the `usr` claim instead of `nik`.
- User/profile contracts expose display name instead of first/last name.
- Auth and user initial migrations describe the new schema. They do not upgrade
  a database that already applied the previous initial migration. Deploy this
  candidate to empty databases; existing databases need separate upgrade migrations.
- Consumers must update together. These candidates are not drop-in replacements
  for the previous stable APIs.

## Service versions

| Component | Candidate |
| --- | --- |
| API gateway | `v0.14.0-rc.1` |
| Auth | `v1.9.0-rc.3` |
| User | `v1.5.0-rc.2` |
| Profile | `v1.4.0-rc.1` |
| Tag | `v1.2.0-rc.2` |
| Recipe | `v1.8.0-rc.2` |
| Encryption | `v1.2.0-rc.2` |
| Shopping list | `v2.5.0-rc.2` |
| Subscription | `v1.1.0-rc.2` |

Auth `rc.2` fixes registration when Firebase is disabled. Auth `rc.3` and the
other database-backed services' `rc.2` use `common/migrate/sql v0.8.1`, which maps
the legacy `pgx` driver name to the registered `pgx5` adapter. API modules remain
at their initial `api/...-rc.1` tags. The service template is a
scaffold, not a deployed service, and has no runtime release in this batch.

Go modules pin published common-library and API versions so individual services
can build with `GOWORK=off`. Repository commits and tags describe source releases;
they do not imply that container images have been published or a VM deployed.
