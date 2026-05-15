#!/bin/bash
# plan-next.sh — On-demand planning session
# Usage: bash scripts/plan-next.sh
# Spawns a sub-agent that reads the spec + current board and proposes the next issues.
# Völundr reviews the output and manually creates approved issues.

set -e

WORKSPACE="/home/omnirdev/.openclaw/workspace"
SPEC="$WORKSPACE/docs/plans/praestos-rewrite-plan.md"
API_KEY="${PAPERCLIP_API_KEY:?PAPERCLIP_API_KEY must be set}"
COMPANY_ID="3adbd3b9-1581-461b-a070-8ae4576d56cf"

echo "=== PraestOS Planning Session ==="
echo "Reading spec and current board state..."
echo ""

# Get current active issues
ACTIVE=$(curl -s "http://127.0.0.1:3100/api/companies/$COMPANY_ID/issues?status=todo,in_progress,in_review" \
  -H "Authorization: Bearer $API_KEY" | python3 -c "
import sys,json
names = {'d6474c23':'Tyr','ec21a603':'Freya','bef9116c':'Heimdall','64dadf00':'Skadi','8fd0b89e':'Völundr'}
issues = json.load(sys.stdin)
out = []
for i in sorted(issues, key=lambda x: x['issueNumber']):
    a = names.get((i.get('assigneeAgentId') or '')[:8],'?')
    out.append(f'  {i[\"identifier\"]} [{i[\"status\"]}] {a}: {i[\"title\"]}')
print('\n'.join(out))
")

echo "=== CURRENT BOARD ==="
echo "$ACTIVE"
echo ""

echo "=== SPEC (phases) ==="
grep "^## Phase\|^### Phase\|^- Phase" "$SPEC" | head -30
echo ""

echo "=== PROPOSAL — next issues to create ==="
echo "Based on the above, Völundr should review and create:"
echo ""

# Parse spec to find next unstarted phase items
python3 << 'PYEOF'
import re

spec = open("/home/omnirdev/.openclaw/workspace/docs/plans/praestos-rewrite-plan.md").read()

# Find Phase 2 section
phase2_match = re.search(r'## Phase 2.*?(?=## Phase 3|\Z)', spec, re.DOTALL)
if phase2_match:
    phase2 = phase2_match.group(0)
    print("PHASE 2 — Multi-tenancy hardening:")
    print(phase2[:1500])
PYEOF
