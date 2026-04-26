# BitSync

A simple utility to sync git repos from your hosted BitBucket, GitHub and GitLab organisations to your local machine.

* The utility can discover the organisations you have access to
* Repos are placed in
  * `$HOME/repos/bitbucket/<bitbucket workspaces>/<projects>/<repos>`
  * `$HOME/repos/github/<github organisations>/<repos>`
  * `$HOME/repos/gitlab/<gitlab groups (incl. nested subgroups)>/<projects>`

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
* In `SSH keys`, upload your SSH public key to enable GIT to authenticate with bitbucket over SSH (use `ssh-keygen` if you don't have one in `~/.ssh/id_rsa.pub` or similar)

Then pick **one** of the API auth methods below:

* **API token (recommended).** From your Atlassian account at `id.atlassian.com`, create an API token with scopes covering account, workspace membership, projects and repositories (read). Use it with `BBEMAIL` (your Atlassian login email) and `BBTOKEN`.
* **App password (legacy, being deprecated).** In `Account settings` make a note of your Bitbucket `username` (usually different to your email). In `App passwords` create one with read rights to account, workspace membership, projects and repositories. Use it with `BBUSER` and `BBAPPPASS`.

If both are set, `BBEMAIL`+`BBTOKEN` wins.

## Configuring Github for access

* In GitHub WebUI click your picture in top right and select `Settings` then `SSH and GPG Keys`
* Click `New SSH key` and upload your public key to enable GIT to authenticate with GitHub over SSH (use `ssh-keygen` if you don't have one in `~/.ssh/id_rsa.pub` or similar)
* If you use SSO to access your GitHub Org, click `Configure SSO` next to your uploaded SSH key and authorise the key for the Org(s) you want to sync
* Scroll down in `settings` to `Developer Settings` and create a `personal access token (classic)`
* Grant the token sufficient rights to access your github orgs/repos and configure SSO if necessary (just like we did for the SSH key)

## Configuring GitLab for access

* In GitLab WebUI click your avatar and select `Edit profile`, then `SSH Keys`
* Add your SSH public key so GIT can authenticate with GitLab over SSH
* Under `Access tokens` (or `Preferences > Access tokens`) create a personal access token with the `read_api` and `read_repository` scopes
* For self-hosted GitLab instances, set `GLURL` to your instance base URL (e.g. `https://gitlab.example.com`); it defaults to `https://gitlab.com`

## Bare mirroring (optional)

If `BITSYNC_MIRROR` is set to `true` then cloning will happen with the `--mirror` git option. This has the effects

* Repo's will be deleted before cloning if they exist
* The repos will be cloned bare with the mirror flag set
* All branches and tags will be mirrored
* Mirrors will be placed under `~/mirrors` instead of `~/repos`

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
# For Bitbucket — pick ONE of these auth pairs
export BBEMAIL=ebeneezer@example.com           # API token auth (recommended)
export BBTOKEN=ATATT3xFfGF0...
# or
export BBUSER=ebeneezer                        # legacy app password auth
export BBAPPPASS=iurfhiuhfIUHFIEUiuehfeuiwF8734Jjhewjfew

export BBORG="nerds-org,jocks-org" # Optional comma seperated list of BB Orgs with no spaces

# For GitHub
export GHTOKEN=ghr-ieufwhiuehfuwehfiuehfuiwhqiuefh
export GHORG="nerds-org,jocks-org" # Optional comma seperated list of GH Orgs/users with no spaces

# For GitLab
export GLTOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
export GLGROUP="nerds-group,nerds-group/sub" # Optional comma seperated list of GL groups/users with no spaces
export GLURL=https://gitlab.com               # Optional, defaults to https://gitlab.com

# Optional: tune concurrency
export BITSYNC_WORKERS=6      # repos per org processed in parallel (default 6)
export BITSYNC_ORG_WORKERS=2  # orgs/workspaces processed in parallel (default 2)

# Now get syncing
bitsync
```
