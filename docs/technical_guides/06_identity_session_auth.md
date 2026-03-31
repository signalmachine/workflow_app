# Identity, Sessions, And Authentication

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how org-scoped identity, browser sessions, bearer sessions, and actor resolution work in `workflow_app`.

## 1. Identity model

The application keeps these concerns distinct:

1. org identity
2. user identity
3. membership role
4. session identity
5. actor identity used by services

The canonical service-level actor is the bridge between auth and business logic.

```go
type Actor struct {
	OrgID     string
	UserID    string
	SessionID string
}
```

That shape is important because downstream services need to know who acted, in what org, and through which session.

## 2. Browser sessions

Browser sessions are password-backed and org-scoped.

The flow is:

1. resolve org slug and email
2. verify password
3. create a session
4. let the HTTP handler set browser cookies from the returned session metadata

Example:

```go
session, err := s.StartBrowserSession(ctx, StartBrowserSessionInput{
	OrgSlug:     input.OrgSlug,
	Email:       input.Email,
	Password:    input.Password,
	DeviceLabel: input.DeviceLabel,
	ExpiresAt:   time.Now().UTC().Add(browserSessionDuration),
})
```

This is the default v1 browser auth path.

## 3. Bearer sessions

The application also supports bearer sessions for non-browser clients on the same backend foundation.

That path is additive and shares the same session truth. It does not create a second auth model.

The important design rule is that browser cookies and bearer tokens are just different transport formats for the same underlying session system.

## 4. Session lifecycle

Session records are durable.

The session layer supports:

1. session creation
2. session inspection
3. session refresh
4. session revocation

The session service is therefore a control boundary, not a convenience helper.

## 5. Bootstrap flow

The repository includes an admin bootstrap path to make first-run access practical.

The bootstrap path upserts the org, user, and membership, then hashes the password.

```go
passwordHash, err := HashPassword(input.Password)
if err != nil {
	return BootstrapAdminResult{}, err
}
```

This is intentionally explicit because the repo is designed to be bootstrapped into a working environment rather than manually seeded with ad hoc SQL.

## 6. Actor resolution

The web layer turns auth into an actor before it calls domain services.

That keeps service-level authorization consistent across browser and API requests.

The general pattern is:

1. read session or header auth
2. resolve org and user identity
3. create `identityaccess.Actor`
4. authorize the transaction or query

## 7. Authorization rules

Roles currently include:

1. admin
2. operator
3. approver

The service layer checks authorization at the transaction boundary. That is important because transport-level authentication alone is not enough to protect workflow actions.

## 8. What to be careful about

Be careful with:

1. session revocation
2. bearer refresh rotation
3. auth-cookie behavior
4. org membership resolution
5. role checks inside transactions
6. keeping header-only compatibility paths from becoming the default

The wrong auth change can silently weaken the control boundary even if the app still appears usable.
