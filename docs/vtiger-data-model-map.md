# Vtiger CE Data Model Map

> Generated from `schema/DatabaseSchema.xml` — 298 tables total
> Repo: https://github.com/SEI-Mkittok/omnir-refactor
> Purpose: Foundation for Omnir CRM modernization / Go backend mapping

---

## Architecture Pattern

Vtiger uses a **split-entity pattern**: every CRM record has:
1. A row in `vtiger_crmentity` — the universal identity, ownership, timestamps, soft-delete
2. One or more **module-specific tables** joined on `crmid = <module>id`

This means every entity is always at minimum a 2-table JOIN.

---

## Core Entity: `vtiger_crmentity`

The base row for every CRM object.

| Column | Type | Notes |
|--------|------|-------|
| `crmid` | INT(19) PK | Universal entity ID |
| `smcreatorid` | INT(19) | Creator user ID |
| `smownerid` | INT(19) | Assigned-to user ID |
| `modifiedby` | INT(19) | Last modified by user ID |
| `setype` | VARCHAR(30) | Module type (e.g. "Accounts", "Contacts") |
| `description` | TEXT | Generic description |
| `createdtime` | TIMESTAMP | Created at |
| `modifiedtime` | TIMESTAMP | Updated at |
| `viewedtime` | TIMESTAMP | Last viewed |
| `status` | VARCHAR(50) | Generic status |
| `version` | INT(19) | Record version |
| `presence` | TINYINT(1) | Visibility flag |
| `deleted` | TINYINT(1) | Soft delete flag |
| `smgroupid` | INT(19) | Assigned group ID |

---

## Module: Accounts

Primary table: `vtiger_account`
Secondary tables: `vtiger_accountbillads`, `vtiger_accountshipads`, `vtiger_accountscf`

| Column | Type | Notes |
|--------|------|-------|
| `accountid` | INT(19) PK | FK → crmentity.crmid |
| `account_no` | VARCHAR(100) | Auto-generated number |
| `accountname` | VARCHAR(100) | Company name |
| `parentid` | INT(19) | Parent account (hierarchy) |
| `account_type` | VARCHAR(200) | Picklist |
| `industry` | VARCHAR(200) | Picklist |
| `annualrevenue` | INT(19) | |
| `rating` | VARCHAR(200) | Picklist |
| `phone` | VARCHAR(30) | |
| `otherphone` | VARCHAR(30) | |
| `email1` | VARCHAR(100) | |
| `email2` | VARCHAR(100) | |
| `website` | VARCHAR(100) | |
| `fax` | VARCHAR(30) | |
| `employees` | INT(10) | |
| `emailoptout` | CHAR(3) | |
| `ownership`, `siccode`, `tickersymbol` | VARCHAR | Financial/corporate fields |

**Bill address** (`vtiger_accountbillads`): street, city, state, country, zip, pobox
**Ship address** (`vtiger_accountshipads`): same structure
**Custom fields** (`vtiger_accountscf`): dynamic extension table

**Key relationships:**
- Has many Contacts (`vtiger_contactdetails.accountid`)
- Has many Potentials (`vtiger_potential.related_to`)
- Has many Quotes, SalesOrders, Invoices (`.accountid`)
- Has many HelpDesk tickets (`vtiger_troubletickets.parent_id`)
- Related to Campaigns via `vtiger_campaignaccountrel`

---

## Module: Contacts

Primary table: `vtiger_contactdetails`
Secondary: `vtiger_contactaddress`, `vtiger_contactsubdetails`, `vtiger_contactscf`, `vtiger_customerdetails`

| Column | Type | Notes |
|--------|------|-------|
| `contactid` | INT(19) PK | FK → crmentity.crmid |
| `contact_no` | VARCHAR(100) | |
| `accountid` | INT(19) | FK → vtiger_account |
| `salutation` | VARCHAR(200) | Picklist |
| `firstname` | VARCHAR(40) | |
| `lastname` | VARCHAR(80) | |
| `email` | VARCHAR(100) | Primary email |
| `phone` | VARCHAR(50) | Office phone |
| `mobile` | VARCHAR(50) | |
| `title` | VARCHAR(50) | |
| `department` | VARCHAR(30) | |
| `otheremail`, `secondaryemail` | VARCHAR(100) | |
| `donotcall`, `emailoptout` | CHAR(3) | Consent flags |

**Sub-details** (`vtiger_contactsubdetails`): homephone, otherphone, assistant, birthday, leadsource
**Address** (`vtiger_contactaddress`): mailing + other address blocks

**Key relationships:**
- Belongs to Account (`accountid`)
- Related to Potentials via `vtiger_contpotentialrel`
- Related to Activities via `vtiger_cntactivityrel`
- Referenced in Quotes, SO, Invoice, PO (`.contactid`)
- Portal access via `vtiger_customerdetails` + `vtiger_portalinfo`

---

## Module: Leads

Primary table: `vtiger_leaddetails`
Secondary: `vtiger_leadaddress`, `vtiger_leadsubdetails`, `vtiger_leadscf`

| Column | Type | Notes |
|--------|------|-------|
| `leadid` | INT(19) PK | FK → crmentity.crmid |
| `lead_no` | VARCHAR(100) | |
| `firstname`, `lastname` | VARCHAR | |
| `salutation` | VARCHAR(200) | |
| `email` | VARCHAR(100) | |
| `company` | VARCHAR(100) | |
| `designation` | VARCHAR(50) | |
| `industry` | VARCHAR(200) | |
| `annualrevenue` | INT(19) | |
| `leadstatus` | VARCHAR(50) | Picklist |
| `leadsource` | VARCHAR(200) | Picklist |
| `converted` | TINYINT(1) | Has been converted to Contact/Account/Potential |
| `rating` | VARCHAR(200) | |
| `noofemployees` | INT(50) | |
| `secondaryemail` | VARCHAR(100) | |

**Address** (`vtiger_leadaddress`): city, state, country, phone, mobile, fax, lane
**Sub-details** (`vtiger_leadsubdetails`): website, donotcall flag, readornot flag, empct

**Conversion mapping**: `vtiger_convertleadmapping` defines how Lead fields map to Contact/Account/Potential on convert.

---

## Module: Potentials (Opportunities)

Primary table: `vtiger_potential`
Secondary: `vtiger_potentialscf`, `vtiger_potstagehistory`

| Column | Type | Notes |
|--------|------|-------|
| `potentialid` | INT(19) PK | FK → crmentity.crmid |
| `potential_no` | VARCHAR(100) | |
| `related_to` | INT(19) | FK → vtiger_account |
| `potentialname` | VARCHAR(120) | Deal name |
| `amount` | DECIMAL(14,2) | |
| `closingdate` | DATE | Expected close |
| `sales_stage` | VARCHAR(200) | Picklist (pipeline stage) |
| `probability` | DECIMAL(7,3) | % |
| `leadsource` | VARCHAR(200) | |
| `nextstep` | VARCHAR(100) | |
| `campaignid` | INT(19) | FK → vtiger_campaign |
| `typeofrevenue`, `potentialtype` | VARCHAR | |

**Stage history**: `vtiger_potstagehistory` — tracks stage changes over time

**Key relationships:**
- Belongs to Account (`related_to`)
- Related to Contacts via `vtiger_contpotentialrel`
- Has many Quotes (`vtiger_quotes.potentialid`)
- Has many SalesOrders (`vtiger_salesorder.potentialid`)

---

## Module: HelpDesk (Tickets)

Primary table: `vtiger_troubletickets`
Secondary: `vtiger_ticketcf`, `vtiger_ticketcomments`

| Column | Type | Notes |
|--------|------|-------|
| `ticketid` | INT(19) PK | FK → crmentity.crmid |
| `ticket_no` | VARCHAR(100) | |
| `title` | VARCHAR(255) | Subject |
| `parent_id` | VARCHAR(100) | FK → account or contact |
| `product_id` | VARCHAR(100) | Related product |
| `priority` | VARCHAR(200) | Picklist |
| `severity` | VARCHAR(200) | Picklist |
| `status` | VARCHAR(200) | Picklist |
| `category` | VARCHAR(200) | Picklist |
| `solution` | TEXT | Resolution |
| `update_log` | TEXT | Change history |
| `hours`, `days` | VARCHAR(200) | Time tracking |

**Comments**: `vtiger_ticketcomments` — threaded comments on tickets
**Related via**: `vtiger_seticketsrel`, `vtiger_salesmanticketrel`

---

## Module: Quotes / Sales Orders / Invoices / Purchase Orders

These share a common inventory pattern. All join `vtiger_inventoryproductrel` for line items.

### `vtiger_quotes`
| Column | Notes |
|--------|-------|
| `quoteid` PK | FK → crmentity.crmid |
| `potentialid` | FK → vtiger_potential |
| `accountid` | FK → vtiger_account |
| `contactid` | FK → vtiger_contactdetails |
| `quotestage` | Picklist status |
| `validtill` | Expiry date |
| `subtotal`, `total`, `adjustment` | Financials |
| `discount_percent`, `discount_amount`, `s_h_amount` | |
| `taxtype` | individual/group |
| `currency_id`, `conversion_rate` | Multi-currency |

### `vtiger_salesorder`
Extends quote → SO flow. Adds `quoteid` (source quote), `sostatus`, `enable_recurring`.

### `vtiger_invoice`
Extends SO → Invoice flow. Adds `salesorderid`, `invoicedate`, `duedate`, `invoicestatus`.

### `vtiger_purchaseorder`
Vendor-side. Links `vendorid`, `contactid`. Has `postatus`.

**Line items** (`vtiger_inventoryproductrel`):
Shared table for all inventory modules. Contains product, qty, unit price, tax, discount per line.

---

## Module: Products

Primary table: `vtiger_products`
Secondary: `vtiger_productcf`, `vtiger_productcurrencyrel`, `vtiger_producttaxrel`, `vtiger_pricebookproductrel`

| Column | Notes |
|--------|-------|
| `productid` INT(11) PK | |
| `productname` | |
| `productcode` | SKU |
| `productcategory` | Picklist |
| `manufacturer` | Picklist |
| `unit_price` | DECIMAL(25,2) |
| `qty_per_unit`, `qtyinstock`, `qtyindemand` | Inventory |
| `reorderlevel` | |
| `taxclass` | FK → vtiger_taxclass |
| `vendor_id` | FK → vtiger_vendor |
| `currency_id`, `conversion_rate` | Multi-currency |

---

## Module: Activities (Calendar/Tasks)

Primary table: `vtiger_activity`
Secondary: `vtiger_activitycf`, `vtiger_recurringevents`, `vtiger_invitees`

| Column | Notes |
|--------|-------|
| `activityid` INT(19) PK | |
| `subject` | |
| `activitytype` | Call / Meeting / Task |
| `date_start`, `due_date` | |
| `time_start`, `time_end` | |
| `status` | Task status |
| `eventstatus` | Event status |
| `priority` | |
| `location` | |
| `recurringtype` | |

**Relations**: `vtiger_seactivityrel` — links activities to any CRM entity

---

## Module: Documents (Notes)

Primary table: `vtiger_notes`

| Column | Notes |
|--------|-------|
| `notesid` INT(19) PK | |
| `title` | |
| `filename` | |
| `notecontent` | TEXT body |
| `folderid` | FK → vtiger_attachmentsfolder |
| `filetype`, `filelocationtype` | |
| `filesize`, `filedownloadcount` | |

**Attachments**: `vtiger_attachments` — binary/file storage metadata
**Relations**: `vtiger_senotesrel` — links docs to any CRM entity

---

## Module: Campaigns

Primary table: `vtiger_campaign`

| Column | Notes |
|--------|-------|
| `campaignid` INT(19) PK | |
| `campaignname` | |
| `campaigntype` | Picklist |
| `campaignstatus` | Picklist |
| `budgetcost`, `actualcost` | |
| `expectedrevenue`, `actualroi` | |
| `numsent` | Emails sent |
| `closingdate` | |

**Membership tables**: `vtiger_campaignaccountrel`, `vtiger_campaigncontrel`, `vtiger_campaignleadrel`

---

## Module: Vendors

Primary table: `vtiger_vendor`
Secondary: `vtiger_vendorcf`

| Column | Notes |
|--------|-------|
| `vendorid` INT(19) PK | FK → crmentity.crmid |
| `vendor_no` | |
| `vendorname` | |
| `phone`, `email`, `website` | |
| `glacct` | GL account |

Related to contacts via `vtiger_vendorcontactrel`.

---

## Users, Roles, Groups

### `vtiger_users`
Core auth + profile table. Notable: `is_admin`, `deleted` (soft), `crypt_type` (MD5 — needs replacing), `accesskey` (API key).

### `vtiger_role`
Hierarchical roles: `roleid`, `rolename`, `parentrole` (hierarchy string), `depth`.

### `vtiger_groups`
User groups for shared ownership. `vtiger_users2group` maps users to groups.

### Access Control
- `vtiger_profile` → permission profiles
- `vtiger_profile2tab`, `vtiger_profile2field` → per-module, per-field permissions
- `vtiger_datashare_*` — 12+ tables controlling inter-role/group data sharing rules
- `vtiger_def_org_share` — org-wide sharing defaults

---

## Workflows

Tables: `com_vtiger_workflows`, `com_vtiger_workflowtasks`, `com_vtiger_workflowtask_queue`

Event-driven automation. Triggers on entity create/update, executes tasks (email, update field, create record, etc).

---

## Meta / System Tables

| Table | Purpose |
|-------|---------|
| `vtiger_tab` | Module registry (enabled modules, order) |
| `vtiger_field` | Field metadata (labels, types, UI config) |
| `vtiger_blocks` | Field grouping blocks for detail view |
| `vtiger_relatedlists` | Subpanel/related list config |
| `vtiger_customview` | Saved views/filters |
| `vtiger_picklist` | Dynamic picklist values |
| `vtiger_modentity_num` | Auto-number sequences (e.g. ACC-001) |
| `vtiger_entityname` | Module display names |
| `vtiger_currency_info` | Multi-currency config |
| `vtiger_ws_*` | Web Services / REST API registry |
| `vtiger_audit_trial` | Audit log |
| `vtiger_tracker` | Recently viewed records |

---

## Key Design Observations for Modernization

1. **Universal join pattern**: Every read needs `JOIN vtiger_crmentity ON crmid = <module>id`. The Go model should embed this automatically.

2. **Soft deletes everywhere**: `vtiger_crmentity.deleted = 1`. Every query needs `WHERE deleted = 0`.

3. **Picklists are dynamic**: Values stored in `vtiger_picklist`-related tables, not as DB enums. The new system should decide: DB enums vs config-driven.

4. **Custom fields (SCF tables)**: Each module has a `_scf` table for admin-added fields. The Go backend needs a strategy here — structured columns vs JSONB extension fields.

5. **Split address blocks**: Accounts have bill/ship (2 separate tables). Contacts have mailing/other in one table. Standardize in new schema.

6. **Passwords in MD5**: `vtiger_users.crypt_type = 'MD5'`. Must migrate to bcrypt/argon2 on import.

7. **No FK enforcement** (historically): MySQL MyISAM era — constraints exist in schema XML but historically weren't enforced. The new PG schema should enforce them properly.

8. **Inventory as a cross-module pattern**: Quotes, SO, Invoice, PO all use the same `vtiger_inventoryproductrel` table. Clean abstraction opportunity.

9. **Permission system is complex**: 12+ datashare tables + profile/field-level permissions. Plan a simplified RBAC model for Omnir.

10. **Web Services layer** (`vtiger_ws_*`): vtiger's legacy REST API is self-describing via these tables. Can be ignored in favor of a native Go API.

---

## Module Inventory (53 modules)

Accounts, Assets, Calendar, Campaigns, Contacts, CustomerPortal, CustomView, Documents, Emails, EmailTemplates, Events, ExtensionStore, Faq, Google, HelpDesk, Home, Import, Install, Inventory, Invoice, Leads, MailManager, Migration, Mobile, ModComments, ModTracker, Oauth2, PBXManager, PickList, Portal, Potentials, PriceBooks, Products, Project, ProjectMilestone, ProjectTask, PurchaseOrder, Quotes, RecycleBin, Reports, Rss, SalesOrder, ServiceContracts, Services, Settings, SMSNotifier, Users, Utilities, Vendors, Vtiger, Webforms, WSAPP

**Core for Omnir MVP**: Accounts, Contacts, Leads, Potentials, HelpDesk, Quotes, SalesOrder, Invoice, Products, Calendar/Activities, Documents, Users, Reports

**Deferrable**: PBXManager, SMSNotifier, Rss, ExtensionStore, Mobile, Oauth2, WSAPP, Google, Portal, Migration
