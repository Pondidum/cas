# Localstorage

This represents the working directory where `cas` is run.  its purpose is to read artifacts from disk, and send to a backend, and also to write the files from the backend to disk.

## `.archive` files

If `ReadFile()` is called with `.archive`, the working directory of the file is compressed into a tar file, and then the resulting archive is returned.

If `WriteFile()` is called with `.archive`, the file is extracted into the directory instead of the `.archive` file being written to disk itself.
