#!/bin/bash
# Wrapper for claude that ensures ANTHROPIC_API_KEY is set and runs non-interactively
export ANTHROPIC_API_KEY="REDACTED_ANTHROPIC_API_KEY"
export PATH="/home/omnirdev/.npm-global/bin:/usr/local/bin:/usr/bin:/bin"
exec /home/omnirdev/.npm-global/bin/claude --dangerously-skip-permissions "$@"
