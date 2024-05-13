# Contributing to Grafana's Shared Workflows Repository

Welcome to the Shared Workflows repository! This repository contains reusable actions designed to streamline and automate common tasks across multiple departments and external organizations. We appreciate your interest in contributing to our shared efforts.

## Scope of `shared-workflows`
This repository is for generic workflows and actions which can be used across multiple types of projects. For more specialised use-cases, we encourage the use of project- or area-specific shared workflow repositories.
Examples which would fit in here:
- Push a container image to a Docker registry
- Run security scanners on built artifacts
- Read and write objects to cloud object storage (GCS or S3)
  And those which would be better in project- or area-specific shared workflow repositories:
- Implement a release process specific to a subset of projects
- Lint project(s) with a single (configurable) toolset

Get in touch with the maintainers [via an issue](https://github.com/grafana/shared-workflows/issues/new) if you're unsure in a particular case.

## Types of Workflows
Unsure about the difference between a reusable workflow and a composite action? Start [here](https://dev.to/n3wt0n/composite-actions-vs-reusable-workflows-what-is-the-difference-github-actions-11kd).

### 1. Reusable Workflows
The `.github/workflows/` directory contains reusable workflows. Refer to the [GitHub documentation](https://docs.github.com/en/actions/using-workflows/reusing-workflows) for more info.

### 2. Composite Actions
The `actions/` directory contains composite actions. Refer to the [GitHub documentation](https://docs.github.com/en/actions/creating-actions/about-custom-actions#composite-actions) for more info.

## Contribution Guidelines

1. **Fork the Repository:** Start by forking the repository to your GitHub account.

2. **Clone the Fork:** Clone your forked repository to your local machine.

3. **Create a Branch:** Create a new branch for your contribution. Use a descriptive name related to the changes you're making.

4. **Make Changes:** Make your desired changes to the codebase.

5. **Document Changes:** Document your changes in a `README.md` file. Try to follow the examples set forward in other Readmes.

6. **Commit Changes:** Commit your changes with clear and concise commit messages.

7. **Push Changes:** Push your changes to your forked repository.

8. **Submit a Pull Request:** Submit a pull request to the main repository. Ensure that your pull request is detailed and includes information about the changes made.

9. **Check CI:** Check your pull request to verify that any CI processes are passing.

10. **Review Process:** Your pull request will be reviewed by Grafana's platform-productivity squad. Make any requested changes or address any feedback during this process.

11. **Contributor License Agreement (CLA):** Before we can accept your pull request, you need to sign [Grafana's CLA](https://grafana.com/docs/grafana/latest/developers/cla/). If you haven't, our CLA assistant prompts you to when you create your pull request.

## Code of Conduct

We uphold a code of conduct to ensure a respectful and inclusive environment for all contributors. Please review Grafana's [Code of Conduct](https://github.com/grafana/grafana/blob/main/CODE_OF_CONDUCT.md) before contributing.

## Contact Us

If you have any questions or need assistance, feel free to [create an issue](https://github.com/grafana/shared-workflows/issues).

Thank you for contributing to our Shared Workflows repository!
