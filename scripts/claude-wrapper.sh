#!/bin/bash
# Wrapper for claude that ensures ANTHROPIC_API_KEY is set and runs non-interactively
export ANTHROPIC_API_KEY="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"
export PATH="/home/omnirdev/.npm-global/bin:/usr/local/bin:/usr/bin:/bin"
exec /home/omnirdev/.npm-global/bin/claude --dangerously-skip-permissions "$@"
