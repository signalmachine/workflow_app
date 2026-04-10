# Browser Operator Getting Started

Date: 2026-04-10
Status: Active
Purpose: explain the first browser path into `workflow_app` and the main pages an operator uses after sign-in.

## 1. Start here

Open one of these browser entry points:

1. `http://127.0.0.1:8080/app/login`
2. `http://127.0.0.1:8080/app`

Use `/app/login` when you want the sign-in form directly. Use `/app` when you want the main app entry point and are fine with the app routing you to sign-in when needed.

## 2. Sign in

Sign in with:

1. org slug
2. user email
3. password
4. device label `browser`

If you used the default local bootstrap values, the org slug is `north-harbor` and the email is `admin@northharbor.local`.

After sign-in, the app should place you on the operator home for that org context.

## 3. What to use next

From the dashboard, move to the task-specific pages:

1. submit a new inbound request from `/app/submit-inbound-request`
2. review parked or in-flight requests from `/app/review/inbound-requests`
3. open a specific request detail from `/app/inbound-requests/{request_reference_or_id}`
4. review processed proposals from `/app/review/proposals`
5. review approvals from `/app/review/approvals`
6. review documents from `/app/review/documents`
7. review accounting from `/app/review/accounting`, then open journal entries, control balances, or tax summaries from that report directory
8. review inventory from `/app/review/inventory`
9. review work orders from `/app/review/work-orders`
10. review audit history from `/app/review/audit`
11. open the operations feed from `/app/operations-feed`
12. open the coordinator-facing browser chat surface from `/app/agent-chat`

## 4. What should stay true

After sign-in:

1. the browser session should remain active until you sign out or the session expires
2. the active org should match the org slug you used during login
3. browser and API reads should agree on the important request, proposal, and approval facts

## 5. Troubleshooting

If you land back on sign-in unexpectedly:

1. confirm the app is running against the database you expected
2. confirm the org slug and email are correct
3. confirm you signed in through the same browser session you are using for the app

If the operator home is not showing the expected workflow links:

1. refresh the page once
2. confirm the session is still active through the sign-in flow
3. check whether the database has the expected bootstrap and workflow records
