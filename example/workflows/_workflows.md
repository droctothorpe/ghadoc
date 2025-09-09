# GitHub Workflows Summary

| Filename | Description | Triggers  |
| --- | --- | --- |
| [add-ci-passed-label.yml](add-ci-passed-label.yml) | Adds the 'ci-passed' label to a pull request once the 'CI Check' workflow completes successfully. | workflow_run |
| [api-server-tests.yml](api-server-tests.yml) | Runs integration tests against API. | push, workflow_dispatch, pull_request |
| [backend-visualization.yml](backend-visualization.yml) | Runs unit tests against backend visualization server. | pull_request, push |
| [build-and-push.yml](build-and-push.yml) | Builds and pushes images to GitHub Container Registry. | workflow_dispatch, workflow_call |
| [e2e-tests.yml](e2e-tests.yml) | Run end-to-end tests against the backend.<br>Validating a multi-line description. | push, pull_request |
| [unit-tests.yml](unit-tests.yml) |  | push, pull_request |
