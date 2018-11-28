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