# BitSync

A simple utility to sync git repos from your hosted BitBucket private organisations to your local machine.

* The utility will discover the organisations you have access to
* Repos are placed in `$HOME/repos/<bitbucket workspaces>/<projects>/<repos>`

## Installing

```bash
go install github.com/trickyearlobe/bitsync@latest
```

## Using

```bash
export BBUSER=ebeneezer
export BBAPPPASS=iurfhiuhfIUHFIEUiuehfeuiwF8734Jjhewjfew
bitsync
```