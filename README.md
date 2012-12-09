# fswatch
## go library for simple UNIX file system watching

`fswatch` runs a watcher in the background to repeatedly check
whether files have been modified. It is capable of watching at least
directories and regular files. If it is watching directories, it
will not see changes for subdirectories.

## License

`fswatch` is licensed under the ISC license.
