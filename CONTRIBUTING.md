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
- Lint project(s) with your team's opinionated toolset

Get in touch with the maintainers [via an issue](https://github.com/grafana/shared-workflows/issues/new) if you're unsure in a particular case.

## Types of Workflows

### 1. Reusable Workflows
Reusable workflows are located in the `.github/workflows/` directory. These workflows are primarily intended for internal use within this repository. They should be used for automating tasks specific to this repository's needs. Reusable workflows are also required if you intend to use the `on: workflow_call` event for workflow triggering. For more information on reusing workflows and using `on: workflow_call`, refer to the [GitHub documentation](https://docs.github.com/en/actions/using-workflows/reusing-workflows).

### 2. Custom Actions
Custom actions are defined in the `actions/` directory. These actions are designed for external contributors to add new functionalities and actions that can be used across various projects. If you have innovative ideas or new actions to contribute, the `actions/` directory is the place to do so. To learn more about creating custom actions, check out the [GitHub documentation](https://docs.github.com/en/actions/creating-actions/about-custom-actions).

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
