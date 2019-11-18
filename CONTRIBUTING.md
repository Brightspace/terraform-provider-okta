# Contributing to Provider

The following is a set of guidelines for contributing to this provider. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Did you find a bug? Or want to propose a change?

If you discovered an issue with the provider you can open up an issue on this repository, using one of the issue templates as a starting point. The issue templates should give you everything you need to get started.

## Contributing

At this time I recommend forking the repository and evolving the provider on your own terms. The providers are used for a narrow slice of automation, to ensure that everything is set up quickly (and to compliance). So I would not expect to see more API coverage from this provider.

### Coding Conventions

Start reading our code and you'll get the hang of it. We optimize for simplicity:

  * We format using `go fmt`
  * We use simple-as-possible variable names (`rule_name` over `super_long_thing_name`)
  * Try to ensure a `main.go` program exists for experimenting with the API
