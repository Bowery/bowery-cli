# CLI

The Bowery CLI (command line interface) is responsible for communicating with the API, Broome, and Delancey in order to manage a developer and their applications.

## Development

First, make sure your local environment is prepared. For more information on this, check out our [installation inscrutions.](https://gist.github.com/sjkaliski/f57f138a93cd81da5a07)

To run the cli

```
$ go get
$ DEBUG=cli ENV=development HOST=10.0.0.15 BROOME_ADDR=10.0.0.14 cli
```

I would recommend placing the above in your path so that you can simply run:

```
$ go get
$ bowery_dev
```

## Testing

Bowery has OS support for:

1. darwin/amd64
2. darwin/386
3. linux/amd64
4. linux/386
5. windows/amd64
6. windows/386

To the best of our abilities, tests need to be run against these platforms.

Using the go builtin testing package, unit tests are expected for new features and changes.

JMoney, our in house deployment bot and all around solid homie, can run integration tests for you too. For now, JMoney is located in the soon to be deprecated [SkyLab](https://github.com/Bowery/SkyLab). Instructions for running JMoney are located there as well.

## Versioning

The CLI abides by semantic versioning (MAJOR.MINOR.PATCH). Bug fixes and the sort increment the third value, new functionality that is backwards compatible increments the second value, and API breaking changes increment the first.

New releases MUST be coordinated with the rest of the engineering team.
