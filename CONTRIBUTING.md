# Contributing to Issuer Node

Welcome to Issuer Node! We're thrilled to have you here. Before you get started, please take a moment to review the following guidelines.

### Contents

- [How to Contribute](#how-to-contribute)
- [Getting Started](#getting-started)
- [Issue Tracker Guidelines](#issue-tracker-guidelines)
- [Code Contribution Guidelines](#code-contribution-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [License Information](#license-information)
- [Contact Information](#contact-information)

## How to Contribute

**Reporting Issues**: If you encounter any bugs or have suggestions for improvements, please open an issue on GitHub. If the bug is a security vulnerability, please report it directly [here](https://support.polygon.technology/support/solutions/categories/82000473421/folders/82000694808).

**Requesting Features**: If you have ideas for new features or enhancements, please submit a feature request on GitHub.

**Submitting Changes**: Fork the repository, make your changes, and submit a pull request. Be sure to follow the guidelines outlined below.

## Getting Started

To set up the project locally, follow the [README](./README.md#quick-start-installation) instructions.

For an advanced setup, visit our [extended documentation](https://devs.polygonid.com/docs/issuer/issuer-configuration).

## Issue Tracker Guidelines

Search for existing issues before creating new ones.

Provide detailed information and steps to reproduce when reporting bugs.

Follow the issue template if available.

## Code Contribution Guidelines

Submit concise and focused pull requests with clear descriptions.

Follow the [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) specification both for the commits and PR names.

Use develop as the base and target branch for pull requests.

Be responsive to feedback and address any review comments promptly.

## Testing Guidelines

Write tests for new features or changes 

Ensure all existing tests pass and the linter reports no errors before submitting your changes.

Run tests and linter locally with:
``` bash
make up-test // To start the database used by tests
make tests // run all tests
make lint // run linter
``` 

## Documentation

Keep documentation up-to-date with any changes or additions.
Help improve existing documentation or contribute new documentation as needed.

## License Information

By contributing to this project, you agree to the terms of licenses [Apache](LICENSE-APACHE) and [Mit](LICENSE-MIT).

## Contact Information

If you have any questions or need assistance, feel free to contact the project maintainers [here](https://support.polygon.technology/support/solutions/categories/82000473421/folders/82000694808).