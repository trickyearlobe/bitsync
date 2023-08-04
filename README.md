# BitSync

A simple utility to sync git repos from your hosted BitBucket private organisations to your local machine.

* The utility will discover the organisations you have access to
* Repos are placed in `$HOME/repos/<bitbucket workspaces>/<projects>/<repos>`

## Installing
First of all, install GO (aka Golang) if you don't alraedy have it.

Then use GO to install BitSync.

```bash
go install github.com/trickyearlobe/bitsync@latest
```

The BitSync binary will be downloaded, compiled and installed to `~/go/bin/bitsync` so make sure it gets added to your path (ideally into a shell startup script like `.bash_profile` or `.zshrc`

```bash
export PATH=$PATH:~/go/bin
```

## Using

The tool clones repos using SSH so you will need to generate and add SSH keys to your bitbucket account.

You will also need to add an `AppPass` to your bitbucket user so that bitsync can access the API. You can do this by logging in to BitBucket, and clicking the gears icon. Make sure you also check your user ID in account settings... it's usually not the eMail address you use to log in.

```bash
export BBUSER=ebeneezer
export BBAPPPASS=iurfhiuhfIUHFIEUiuehfeuiwF8734Jjhewjfew
bitsync
```
