The monom file is configured in the root of a monomn project. Internally the monom config file is aliased with the `$MONOM_CONFIG_FILE` variable.

The monom config file exposes an api of these functions:

1. `run` - accepts stdin of the command the user has entered. Print to stdout a full path to the file to be executed.
1. `complete` - Print to stdout all possible file paths that user can run.

