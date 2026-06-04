# Phase 7 Review

## Findings

1. **P1 - CI YAML syntax check depends on PyYAML but does not install it.**  
   `.github/workflows/ci.yml:29-48` runs `python3 -c "import yaml ..."`. `yaml` is not part of Python's standard library, and this repository does not install `PyYAML` before that step. I confirmed locally that `python3 -c 'import yaml'` fails with `ModuleNotFoundError: No module named 'yaml'`. On a clean GitHub-hosted runner this can make every CI run fail before it reaches the useful checks. Use the Ruby YAML parser from the Phase 7 plan, install `pyyaml` explicitly, or use a checked-in Go/Ruby/Python dependency path that is guaranteed in CI.

2. **P2 - `make ci-check` does not actually match the CI checks.**  
   `Makefile:129-136` says `ci-check` is the same as GitHub Actions, but it omits the YAML syntax check that exists in `.github/workflows/ci.yml:29-48`. That means a developer can run `make ci-check` locally, get a pass, and still fail CI on Kubernetes/app YAML. Add the same YAML parse command to `ci-check`.

3. **P2 - Remote K8s docs still point at the old `phase5` image tag.**  
   `deploy/k8s/README.md:71-83` still says the default image is `ghcr.io/wangchongv8/go-svc-<service>:phase5` and shows the old build/push flow. Phase 7 changed `IMAGE_TAG ?= local` (`Makefile:68-69`) and the current manifests use `phase6`, while releases should prefer Git SHA tags from GitHub Actions. This doc will confuse the target-machine workflow. Update it to explain the new GitHub Actions/GHCR SHA tag flow and `IMAGE_TAG=<git-sha> make k8s-set-images`.

4. **P3 - Code generation in CI is unpinned.**  
   `.github/workflows/ci.yml:49-54` installs `github.com/zeromicro/go-zero/tools/goctl@latest` and then checks generated-code drift. Because `latest` can change independently of this repo, CI may start failing after a goctl release even when project code did not change. Pin goctl to the project version, for example `@v1.10.1`, or document that generator updates are intentionally accepted.

5. **P2 - Manual image builds from feature branches overwrite the `main` image tag.**  
   `.github/workflows/images.yml:3-6` supports `workflow_dispatch`, which is useful for building a feature branch for remote testing. However `.github/workflows/images.yml:57-59` always pushes both `${{ github.sha }}` and `main`. If a developer manually runs the workflow on a feature branch, the feature-branch image will overwrite `ghcr.io/wangchongv8/go-svc-<service>:main`. Push the immutable SHA tag for every run, but only push the moving `main` tag when `github.ref == 'refs/heads/main'`. Optionally add a sanitized branch tag for manual feature-branch testing.

## What Looks Good

- `.github/workflows/images.yml:8-10` uses `packages: write` and `GITHUB_TOKEN`, with no hardcoded token.
- `.github/workflows/images.yml:15-59` covers all 5 services with explicit `main` and `conf_dir`.
- Image tags include both immutable `${{ github.sha }}` and moving `main`.
- `IMAGE_TAG` is no longer phase-bound in the Makefile.
- `k8s-set-images` and `k8s-rollout-status` were added.
- The implementation does not introduce direct GitHub Actions deployment to the remote cluster.

## Verification

- `make ci-check` passed locally.
- `git diff --check` passed.
- `bash -n scripts/*.sh` passed.
- `docker compose -f deploy/docker-compose/docker-compose.yml config -q` passed.
- Ruby YAML parsing for `deploy/k8s/*.yaml` and `apps/*/etc/*.yaml` passed.
- `python3 -c 'import yaml'` failed locally with `ModuleNotFoundError`, confirming the workflow dependency risk.
- No Compose/K8s services were started. Port check for `3000,4317,5432,6060,8080,9000-9003,9090,16686,19090` returned no listeners.

## Suggested Fix Order

1. Replace or install the CI YAML parser dependency.
2. Add the same YAML check to `make ci-check`.
3. Update `deploy/k8s/README.md` to remove `phase5` and document the Phase 7 SHA-tag release flow.
4. Prevent feature-branch `workflow_dispatch` image builds from pushing the `main` tag.
5. Pin `goctl` in CI.
