if (-not $env:PAPERCLIP_API_KEY) {
    throw "PAPERCLIP_API_KEY must be set in the local environment."
}

$headers = @{"Authorization" = "Bearer $env:PAPERCLIP_API_KEY"; "Content-Type" = "application/json"}
$baseUrl = "http://100.109.245.95:3100/api/issues"

function Update-Issue($id, $assignee, $spec) {
    try {
        $issue = Invoke-RestMethod -Uri "$baseUrl/$id" -Headers $headers -Method Get
        $newDesc = $issue.description + "`n`n## Technical Specification (Völundr)`n" + $spec
        $body = @{
            description = $newDesc
            assigneeAgentId = $assignee
            status = "todo"
        } | ConvertTo-Json
        Invoke-RestMethod -Uri "$baseUrl/$id" -Headers $headers -Method Patch -Body $body | Out-Null
        Write-Output "Successfully updated $id"
    } catch {
        Write-Output "Failed to update $id"
    }
}

$tyr = "d6474c23-67e2-443f-9b6e-84a3e266aa3d"
$freya = "ec21a603-d783-40ba-9ea5-cd0038eae05a"
$skadi = "64dadf00-a32f-4edb-a964-30380cfd00f1"

Update-Issue "OMN-576" $tyr "**Backend (Tyr):** Implement CRUD for email_templates (POST/GET/PATCH/DELETE /api/v1/email-templates). Update Sequence Steps to accept template_id. `n**Frontend (Freya):** Integrate WYSIWYG editor (TipTap). Update sequence builder to select templates."
Update-Issue "OMN-575" $tyr "**Frontend (Freya):** Add related entity combo-boxes (Account, Contact, Pipeline) to Quote form. `n**Backend (Tyr):** Investigate POST /api/v1/quotes 400 error on drafts. Fix validation."
Update-Issue "OMN-574" $freya "**Frontend (Freya):** Update columns definition in Contacts Table. Ensure accessor mapping matches account.name correctly."
Update-Issue "OMN-573" $freya "**Frontend (Freya):** Update Deal Card component in Pipeline Kanban view. Render account.name, contact.name, and format expected_close_date."
Update-Issue "OMN-572" $tyr "**Backend (Tyr):** Check billing/entitlement middleware returning 402 on GET /api/v1/contacts/:id. Ensure read access isn't blocked erroneously."
Update-Issue "OMN-571" $freya "**Frontend (Freya):** Add inline editing to Deal Value field on Deal Detail panel -> PATCH /api/v1/deals/:id."
Update-Issue "OMN-570" $freya "**Frontend (Freya):** Audit summary cards and detail panels. Wrap Account and Contact name spans in React Router <Link> components pointing to their detail pages."
Update-Issue "OMN-569" $skadi "**QA (Skadi):** Execute E2E testing for email thread mark-as-read based on OMN-552."
