# BitSync

A simple utility to sync git repos from your hosted BitBucket and GitHub private organisations to your local machine.

* The utility can discover the organisations you have access to
* Repos are placed in
  * `$HOME/repos/bitbucket/<bitbucket workspaces>/<projects>/<repos>`
  * `$HOME/repos/github/<github organisations>/<repos>`

## Installing
First of all, install GO (aka Golang) if you don't already have it.

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

## Configuring Github for access

* In GitHub WebUI click your picture in top right and select `Settings` then `SSH and GPG Keys`
* Click `New SSH key` and upload your public key to enable GIT to authenticate with GitHub over SSH (use `ssh-keygen` if you don't have one in `~/.ssh/id_rsa.pub` or similar)
* If you use SSO to access your GitHub Org, click `Configure SSO` next to your uploaded SSH key and authorise the key for the Org(s) you want to sync
* Scroll down in `settings` to `Developer Settings` and create a `personal access token (classic)`
* Grant the token sufficient rights to access your github orgs/repos and configure SSO if necessary (just like we did for the SSH key)

## Bare mirroring (optional)

If `BITSYNC_MIRROR` is set to `true` then cloning will happen with the `--mirror` git option. This has the effects

* Repo's will be deleted if they exist
* The repos will be cloned bare with the mirror flag set
* All branches and tags will be mirrored

This is effectively a full archive of the repo, but it cannot be used for normal git workflows as it has the following git config options set

```aiignore
[core]
	bare = true

[remote "origin"]
	fetch = +refs/*:refs/*
	mirror = true
```

## Using

Pass the credentials, and optional org lists, as environment variables and sync your repos.

```bash
# For Bitbucket
export BBUSER=ebeneezer
export BBAPPPASS=iurfhiuhfIUHFIEUiuehfeuiwF8734Jjhewjfew
export BBORG="nerds-org,jocks-org" # Optional comma seperated list of BB Orgs with no spaces

# For GitHub
export GHTOKEN=ghr-ieufwhiuehfuwehfiuehfuiwhqiuefh
export GHORG="nerds-org,jocks-org" # Optional comma seperated list of GH Orgs with no spaces

# Now get syncing
bitsync
```
