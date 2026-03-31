# Browser Sign-In And Admin Bootstrap

Date: 2026-03-30
Status: Active
Purpose: explain the recommended best-practice path for creating a real local admin login instead of relying on generated verification identities.

## 1. Why this exists

The repository includes verification and integration paths that generate timestamp-based org slugs and email addresses so repeated automated runs do not collide with each other.

Those generated identities are appropriate for:

1. tests
2. verification runs
3. disposable seeded data

They are not a good default for a human operator login.

## 2. Recommended local admin path

Use the dedicated bootstrap command:

```bash
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
```

This is preferable to:

1. looking up random generated user IDs by hand
2. manually editing password hashes
3. using verification-created orgs and users as the long-term browser login

The browser sign-in flow is available at `/app/login`, and `/app` routes unauthenticated visitors to the same sign-in experience.

## 3. What the bootstrap command does

`cmd/bootstrap-admin` applies the following best-practice setup shape against the main application database:

1. ensure one friendly org slug exists
2. ensure one friendly admin user exists
3. ensure the user has an active membership in that org
4. ensure the membership role is `admin`
5. hash and store the password through the same identity layer used by sign-in

Plain-language meaning:

1. `bootstrap` means setting up the minimum initial records needed so the application can be used
2. it is the first-run setup path for a friendly organization and admin login
3. it does not create a second app or a separate isolated environment
4. it creates or reuses records inside the main database pointed to by `DATABASE_URL`

Name and identifier distinction:

1. org slug is the short sign-in identifier, for example `north-harbor`
2. org name is the human-readable organization label, for example `North Harbor Works`
3. org ID is the internal UUID stored in the database and is not the value used on the browser sign-in page

The command is safe to rerun when you need to:

1. recreate a local login on a fresh database
2. standardize a friendlier local org slug and email
3. rotate the bootstrap admin password

## 4. Default example

With the default values, use:

1. org slug `north-harbor`
2. email `admin@northharbor.local`
3. password: the password you passed on the command line

Example:

```bash
go run ./cmd/bootstrap-admin -password 'NorthHarbor2026'
go run ./cmd/app
```

Then sign in with:

1. Org slug: `north-harbor`
2. User email: `admin@northharbor.local`
3. Password: `NorthHarbor2026`
4. Device label: `browser`

If you use those defaults, you are logging into the organization `North Harbor Works`. That org is one tenant inside the application’s main database, and your session becomes active in that org context after sign-in.

## 5. Production-minded guidance

Even in local development, keep the setup path aligned with common engineering practice:

1. do not hardcode default passwords in migrations
2. do not auto-create magic admin users every time the app starts
3. keep the test database and main application database separate
4. prefer one explicit bootstrap step over hidden startup-side effects
5. rotate the bootstrap password when sharing an environment

That separation keeps local setup convenient without weakening the application’s control boundary.
