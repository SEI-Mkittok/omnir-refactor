
## SSH to omnir-dev-2 (Staging)

Use key-based SSH — no Tailscale auth prompt:
```bash
ssh -i ~/.ssh/omnir_deploy omnirdev@100.73.134.90
```

Do NOT use `ssh omnirdev@100.73.134.90` or `ssh omnir-dev-2` — those require browser auth approval.

