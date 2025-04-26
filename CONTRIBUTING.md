# Contributing to Umono
We’re excited that you’re interested in contributing to Umono! We believe that open source thrives when people work together to build something useful for everyone. Here’s a guide to help you get started with contributing.

## How to Contribute
There are several ways you can contribute to the Umono project:

### 1. Reporting Bugs or Issues

If you’ve found a bug or have an issue to report, please [open an issue](https://github.com/umono-cms/umono/issues/new). Provide as much detail as possible, including:
- Steps to reproduce the issue
- The expected and actual behavior
- Screenshots or logs, if applicable

### 2. Submitting Code
If you'd like to contribute code, follow these steps:

#### 1. Open an issue first.
Before submitting a pull request, please open an issue describing what you plan to work on. This helps us ensure that the feature or bug fix is aligned with the project goals.

#### 2. Fork the repository
Fork the repository to your GitHub account.

#### 3. Create a new branch
Create a new branch for your feature or bug fix (e.g., `feature/my-new-feature` or `fix/issue-123`).

#### 4. Make your changes
Implement the feature or fix, following the coding standards and ensuring that your code is clean, understandable, and reusable.

#### 5. Write tests
It is mandatory to write tests for any new features or bug fixes. This ensures that the code works as expected and helps maintain the stability of the project. If you find that existing tests need to be updated, improved, or fixed, please feel free to adjust them as necessary. Don't hesitate to improve test coverage or fix any issues in the existing test code. This is crucial for maintaining the reliability of the project.

#### 6. Push your changes and create a pull request
Push your changes and create a pull request to the main branch. Include a clear description of what your changes do and why they are needed.

#### 7. Working with Multiple Repositories
If your contribution involves changes to the [backend](https://github.com/umono-cms/umono), [admin UI](https://github.com/umono-cms/admin-ui), or the Umono language repository ([umono-lang](https://github.com/umono-cms/umono-lang)), please ensure that all relevant repositories are updated appropriately. Here's what to do:
##### 1. Update the Relevant Repository:
If your changes only affect the backend, admin UI, or umono-lang, open an issue directly in the respective repository. For example:
- **Backend-related bugs**: [Open an issue](https://github.com/umono-cms/umono/issues/new) in the main repository.
- **Admin UI-related bugs**: [Open an issue](https://github.com/umono-cms/admin-ui/issues/new) in the admin-ui repository.
- **umono-lang-related bugs**: [Open an issue](https://github.com/umono-cms/umono-lang/issues/new) in the umono-lang repository.

##### 2. Create Separate Pull Requests for Each Repository:
If your changes affect more than one repository, create separate pull requests for each one, ensuring they are in sync.

##### 3. Follow the Same Guidelines for All Repositories:
Make sure you follow the contribution guidelines for each repository, including writing tests, creating issues, and submitting pull requests.

##### 4. Communicate Changes Between Repositories:
If there are dependencies between the changes in the backend, admin UI, or umono-lang, be sure to describe them clearly in the pull request descriptions. This will help reviewers understand how the changes are related across repositories.

##### 5. Stay Synchronized:
When working with multiple repositories, it's important to keep your work synchronized. If you update one repository, ensure that the other repositories are also updated as needed.

#### 8. Keep It Simple (KISS Principle)
We follow the Keep It Simple, Stupid (KISS) principle in Umono. When proposing new features or fixes, try to keep your solutions as simple as possible. Avoid overcomplicating things—our goal is to build something that is easy to use and maintain. This will help us keep the project lightweight and accessible to everyone.


## Additional Guidelines
- **Write clear, concise commit messages.** This helps reviewers understand the context and changes you're making.
- **Document your code.** If you add a new feature, include comments.
- **Test your code.** Ensure that everything works as expected and that tests cover all new changes.

## Transparency
As project maintainers, we sometimes contribute to the project without writing tests for small or quick changes. This is done to ensure that rapid development can occur. However, we strongly encourage all contributors to write tests for any new features or bug fixes, as this helps maintain the stability and reliability of the project in the long run.

## Need Help?
If you’re new to open source or need help getting started, feel free to reach out by opening an issue or asking for assistance in the GitHub discussions. We’re happy to help!
