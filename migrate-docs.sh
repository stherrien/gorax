#!/bin/bash

# Migrate old project notes to .context folder

echo "📦 Migrating project notes to .context/project-notes..."

# Create .context directory if it doesn't exist
mkdir -p .context/project-notes

# Move all old project note files (ALL_CAPS .md files)
mv docs/AUTHENTICATION.md .context/project-notes/
mv docs/REQUIREMENTS.md .context/project-notes/
mv docs/SECRETS_MANAGER_DESIGN.md .context/project-notes/
mv docs/multi-tenancy-phase-1.4.md .context/project-notes/
mv docs/phase-1.10-websocket-implementation.md .context/project-notes/
mv docs/SLACK_SENDDM_IMPLEMENTATION.md .context/project-notes/
mv docs/SLACK_INTEGRATION_DESIGN.md .context/project-notes/
mv docs/CREDENTIAL_USAGE.md .context/project-notes/
mv docs/websocket-api.md .context/project-notes/
mv docs/QUICKSTART_AUTH.md .context/project-notes/
mv docs/PHASE_2.6_IMPLEMENTATION_SUMMARY.md .context/project-notes/
mv docs/QUEUE_INTEGRATION.md .context/project-notes/
mv docs/CREDENTIAL_E2E_TEST_RESULTS.md .context/project-notes/
mv docs/TASKS.md .context/project-notes/
mv docs/OAUTH_SETUP.md .context/project-notes/
mv docs/PHASE_1_3_SUMMARY.md .context/project-notes/
mv docs/PROJECT_STATUS_2025-12-17.md .context/project-notes/
mv docs/SLACK_UPDATEMESSAGE_IMPLEMENTATION.md .context/project-notes/
mv docs/PHASE_2.1_SUMMARY.md .context/project-notes/
mv docs/TEST_PLAN_API_INTEGRATION.md .context/project-notes/
mv docs/SLACK_SENDMESSAGE_IMPLEMENTATION.md .context/project-notes/
mv docs/PHASE_3_2_SUMMARY.md .context/project-notes/
mv docs/CONDITIONAL_ACTIONS.md .context/project-notes/
mv docs/executor-implementation.md .context/project-notes/

echo "✅ Migration complete!"
echo ""
echo "Project notes are now in: .context/project-notes/"
echo "User documentation is in: docs/"
echo ""
echo "Files in docs/:"
ls -1 docs/
