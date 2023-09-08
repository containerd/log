# log

The log package provides a common logging interface across containerd repositories and a way for clients to use and configure logging in containerd packages.

This package is not intended to be used as a standalone logging package outside of the containerd ecosystem and is intended as an interface wrapper around a logging implementation.
In the future this package may be replaced with a common go logging interface.
