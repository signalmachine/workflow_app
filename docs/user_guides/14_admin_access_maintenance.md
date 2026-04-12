# Admin Access Maintenance

Date: 2026-04-12
Status: Active
Purpose: explain how an admin maintains org-scoped user access and membership roles.

## 1. Open access maintenance

Use an admin session and open:

1. `/app/admin`
2. `/app/admin/lists`
3. `/app/admin/access`

Use this page for org membership and role maintenance. It is not a separate identity product; it controls who can use the current org's workflow surfaces.

## 2. Add or attach a user

Use `/app/admin/access` to create or attach an org membership for a real user email.

Example:

North Harbor Works adds `reviewer@northharbor.local` so another operator can review approvals during user testing. The admin creates or attaches the user on `/app/admin/access`, assigns the intended role, and confirms the membership appears in the list before sharing sign-in instructions.

## 3. Update a role

Use the membership role action when a user's responsibility changes.

Example:

An operator who only reviewed requests now needs to maintain master data for a test session. An admin updates the membership role to admin, then asks the operator to refresh the browser session and confirm `/app/admin` is reachable.

## 4. Guardrails

Keep these controls intact:

1. do not remove the only admin path needed to maintain the org
2. do not use shared passwords for separate operators
3. do not create test users in the wrong org
4. confirm access changes through browser sign-in rather than assuming the list update is enough

## 5. Troubleshooting

If a user cannot reach an admin route:

1. confirm the user signed in with the correct org slug
2. confirm the email matches the membership record
3. confirm the membership role on `/app/admin/access`
4. sign out and sign in again if the session was started before the role changed
