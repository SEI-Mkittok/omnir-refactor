#!/bin/bash
# Wrapper for claude that ensures ANTHROPIC_API_KEY is set and runs non-interactively.
# Keep the real key in the environment or a local-only wrapper, never in Git.
: "${ANTHROPIC_API_KEY:?ANTHROPIC_API_KEY must be set}"
export PATH="/home/omnirdev/.npm-global/bin:/usr/local/bin:/usr/bin:/bin"
exec /home/omnirdev/.npm-global/bin/claude --dangerously-skip-permissions "$@"
