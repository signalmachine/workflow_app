# Admin Party Setup

Date: 2026-04-12
Status: Active
Purpose: explain how an admin maintains bounded customer and vendor support records for workflow use.

## 1. Open party setup

Use an admin session and open:

1. `/app/admin`
2. `/app/admin/master-data`
3. `/app/admin/parties`
4. `/app/admin/parties/{party_id}`

Use party setup for customer and vendor support records. Do not turn it into a broad CRM workflow; records here exist to support request, document, accounting, inventory, and work-order continuity.

## 2. Create a party

Use `/app/admin/parties` to create a bounded customer or vendor record.

Example:

Before testing a service-invoice workflow, an admin creates customer `Harbor Retail Pvt Ltd` with customer type and a primary contact. The later invoice request should reference that same party instead of inventing a duplicate customer name inside the request payload.

## 3. Review exact party detail

Open `/app/admin/parties/{party_id}` when you need to confirm contact data or status before a downstream workflow uses the party.

Example:

A processed proposal points to a vendor but the reviewer is unsure whether the vendor contact is current. Open the exact party detail from admin party setup, confirm the contact, then return to the proposal or document review chain.

## 4. Govern active status

Use active or inactive status controls when a party should stop being used for new workflow activity while preserving historical linkage.

Example:

If a vendor is no longer valid for new purchase workflows, mark the vendor inactive. Existing documents and audit trails should still trace to the same party record.

## 5. Troubleshooting

If a party is missing downstream:

1. check `/app/admin/parties`
2. confirm the party was created in the current org
3. confirm it is active when the downstream task requires active parties
4. avoid creating a duplicate until the exact party detail has been checked
