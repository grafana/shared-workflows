# Migration Guide

This guide explains how to migrate an existing repository (like grafana-com) to use the shared-workflows `component-change-detection` action.

## Prerequisites

Before migrating, ensure:
1. This action has been merged to `main` in `grafana/shared-workflows`
2. Your repository has a `.component-deps.yaml` configuration file
3. Your deployment workflow uploads `component-tags` artifacts

## Migration Steps for grafana-com

### Step 1: Update `.github/workflows/build.yml`

Replace the local action with the shared action:

**Before:**
```yaml
- name: Detect changed components
  id: detect-changes
  uses: ./.github/actions/detect-changed-components
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
```

**After:**
```yaml
- name: Detect changed components
  id: detect-changes
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
```

### Step 2: Update `.github/workflows/deploy-prod.yml`

Same change as above:

```yaml
- name: Detect changed components
  id: detect-changes
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
```

### Step 3: Remove Old Implementation

Once the workflows are updated and tested, remove the old implementation:

```bash
# Remove local actions
rm -rf .github/actions/detect-changed-components
rm -rf .github/actions/download-previous-component-tags

# Remove Go tool
rm -rf tools/change-detection

# Remove scripts (if not used for other purposes)
rm scripts/detect-changes

# Remove Makefile targets
# Edit Makefile to remove:
# - bin/changed-components target
# - clean-changed-components target
```

### Step 4: Update Makefile (if applicable)

Remove the change detection build targets:

```makefile
# Remove these lines:
bin/changed-components: tools/change-detection/cmd/changed-components/main.go $(shell find tools/change-detection/pkg/changedetector -name "*.go" 2>/dev/null)
	@mkdir -p bin
	cd tools/change-detection && go build -o ../../$@ ./cmd/changed-components

.PHONY: clean-changed-components
clean-changed-components:
	rm -f bin/changed-components
```

## Migration Steps for Other Repositories

### For Repositories Starting Fresh

1. **Create `.component-deps.yaml`** in your repository root:

```yaml
global_excludes:
  - '**/*.test.*'
  - 'docs/**'
  - '**/*.md'

components:
  your_component:
    paths:
      - 'src/**'
    excludes: []
    dependencies: []
```

2. **Add change detection to your build workflow**:

```yaml
jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      changes_json: ${{ steps.detect.outputs.changes_json }}
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          fetch-depth: 100

      - name: Detect changes
        id: detect
        uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
        with:
          config-file: '.component-deps.yaml'
          previous-tags-source: 'deploy.yml'  # Your deployment workflow

  build:
    needs: detect-changes
    runs-on: ubuntu-latest
    steps:
      - name: Build if changed
        if: fromJSON(needs.detect-changes.outputs.changes_json).your_component == true
        run: make build
```

3. **Update your deployment workflow** to save component tags:

```yaml
jobs:
  deploy:
    steps:
      # ... deployment steps ...
      
      - name: Save component tags
        run: |
          # Create component-tags.json with format:
          # {"component": {"commitSHA": "...", "digest": "..."}}
          cat > component-tags.json <<EOF
          {
            "your_component": {
              "commitSHA": "${{ github.sha }}",
              "digest": "${{ steps.build.outputs.digest }}"
            }
          }
          EOF
      
      - name: Upload component tags
        uses: actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6.0.0
        with:
          name: component-tags
          path: component-tags.json
          retention-days: 90
```

## Testing the Migration

### Before Merging

1. **Test locally** (if possible):
   ```bash
   # Build the tool
   cd actions/component-change-detection
   go build -o changed-components ./cmd/changed-components
   
   # Create a test tags file
   echo '{"component1": "abc123"}' > /tmp/test-tags.json
   
   # Run detection
   ./changed-components \
     --config ../../grafana-com/.component-deps.yaml \
     --tags /tmp/test-tags.json \
     --verbose
   ```

2. **Create a test PR** in grafana-com that:
   - Updates workflows to use the new action
   - Makes a small change to verify detection works
   - Checks the workflow logs to confirm correct behavior

### After Merging to shared-workflows

1. Create a PR in grafana-com with workflow updates
2. Verify the change detection step succeeds
3. Verify components are correctly identified as changed/unchanged
4. Merge the grafana-com PR
5. Clean up old implementation in a follow-up PR

## Rollback Plan

If issues are discovered after migration:

1. **Immediate rollback**: Revert the workflow changes to use the local action
2. **Fix forward**: Update the shared-workflows action and pin to a specific commit/tag

## Version Pinning (Recommended for Production)

Instead of using `@main`, pin to a specific version:

```yaml
uses: grafana/shared-workflows/actions/component-change-detection@v1.0.0
```

This provides:
- Stability (no surprise changes)
- Easier rollback (revert to previous version)
- Better audit trail (know exactly what version was used)

## Support

For issues or questions:
- File an issue in [grafana/shared-workflows](https://github.com/grafana/shared-workflows/issues)
- Check the [README](./README.md) for troubleshooting tips
- Contact the Stack State Service team on Slack: #stack-state-service

## Checklist

- [ ] Merged component-change-detection action to grafana/shared-workflows
- [ ] Updated build workflow to use shared action
- [ ] Updated deployment workflow to use shared action  
- [ ] Tested in a PR that changes are detected correctly
- [ ] Merged workflow updates to main
- [ ] Removed old local implementation
- [ ] Updated Makefile (if applicable)
- [ ] Updated documentation/README references (if any)
