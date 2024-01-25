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

The BitSync binary will be downloaded, compiled and installed to `~/go/bin/bitsync` so make sure it gets added to your path, ideally into a shell startup script like `.bash_profile` or `.zshrc`

```bash
export PATH=$PATH:~/go/bin
```

Finally, make sure you have an up to date command line version of `git` installed on your PATH (this app shell's out to it)

## Configuring Bitbucket for access

* In BitBucket WebUI click the gear icon and select `Personal Bitbucket settings`
* In `Account settings`, make a note of your `username` under Bitbucket Profile Settings. It is usually different to your eMail.
* In `SSH keys`, upload your SSH public key to enable GIT to authenticate with bitbucket over SSH (use `ssh-keygen` if you don't have one in `~/.ssh/id_rsa.pub` or similar)
* In `App passwords` create and copy an App password which has read rights to account, workspace membership, projects and repositories so that we can autodiscover your repos.

## Using

Pass the credentials as environment variables and sync your repos.

```bash
export BBUSER=ebeneezer
export BBAPPPASS=iurfhiuhfIUHFIEUiuehfeuiwF8734Jjhewjfew
bitsync
```
