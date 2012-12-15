# fswatch
## go library for simple UNIX file system watching

`fswatch` runs a watcher in the background to repeatedly check
whether files have been modified. It is capable of watching at least
directories and regular files. If it is watching directories, it
will not see changes for subdirectories.

## Usage
There are two types of Watchers:

* static watchers watch a limited set of files; they do not purge deleted
files from the watch list.
* auto watchers watch a set of files and directories; directories are
watched for new files. New files are automatically added, and deleted
files are removed from the watch list.

Take a look at the provided `clinotify/clinotify.go` for an example.

## License

`fswatch` is licensed under the ISC license.
