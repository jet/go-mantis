NOTICE: SUPPORT FOR THIS PROJECT ENDED ON 18 November 2020

This projected was owned and maintained by Jet.com (Walmart). This project has reached its end of life and Walmart no longer supports this project.

We will no longer be monitoring the issues for this project or reviewing pull requests. You are free to continue using this project under the license terms or forks of this project at your own risk. This project is no longer subject to Jet.com/Walmart's bug bounty program or other security monitoring.


## Actions you can take

We recommend you take the following action:

  * Review any configuration files used for build automation and make appropriate updates to remove or replace this project
  * Notify other members of your team and/or organization of this change
  * Notify your security team to help you evaluate alternative options

## Forking and transition of ownership

For [security reasons](https://www.theregister.co.uk/2018/11/26/npm_repo_bitcoin_stealer/), Walmart does not transfer the ownership of our primary repos on Github or other platforms to other individuals/organizations. Further, we do not transfer ownership of packages for public package management systems.

If you would like to fork this package and continue development, you should choose a new name for the project and create your own packages, build automation, etc.

Please review the licensing terms of this project, which continue to be in effect even after decommission.

ORIGINAL README BELOW

----------------------

# Mantis

A collection of libraries implementing some patterns and functionality that is common across many applications.  You can sort of think about Mantis as a "standard library" for Jet's Golang codebase.

The goal of Mantis is two fold:

- Develop a suite of high-level behaviors which can be implemented by first or third-party libraries.
- Implement some of the core functionality not present in the go standard library which is common across many applications and libraries - without needing to import a lot of third-party libraries which may be yanked at any time, or cause import problems due to renaming.

## Contributing to Mantis

Libraries in Mantis should strive to be:

- **Well tested**: Each library should have a reasonably high test coverage score where warrented. Bug fixes should have tests to ensure they do not regress between releases.
- **Well documented**: Godoc comments on all exported types and functions. No Godoc place-holders just to make linters happy. If it can't be documented, consider not exporting it.
- **Zero Dependencies***: Depend on few (preferably zero) external dependencies outside the go standard library. Prefer creating unexported interfaces to allow 3rd-party libraries to be used as needed. (*See "Dependency Exceptions" for a list of 3rd-party dependencies that are used and rationale.)

## Dependency Exceptions

This section lists the third-party libraries that Mantis depends on which exists outside the go standard library or Mantis. The goal is to minimize this list to zero if possible. Where not possible, the library choice should favor the most well tested library that has minimal (preferably zero) transitive dependencies beyond the standard library.

- [`github.com/pkg/errors`](https://github.com/pkg/errors) - A well tested and maintained library by [Dave Cheney](https://dave.cheney.net/) (and [others](https://github.com/pkg/errors/graphs/contributors)) for providing more contextual errors including cause and stacktrace information. This is a good candidate for inclusion because:
    1. Non-trivial to re-implement.
    2. Well maintained and has 100% test coverage.
    3. No transitive dependencies beyond the go standard library.
    4. Types introduced in this package are hermetically sealed behind the standard library interface `errors` - so there is also no leakage of implementation.
