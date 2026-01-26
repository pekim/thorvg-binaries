# thorvg-binaries

This repo provides the means to update the thorgv shared libraries of
https://github.com/pekim/thorvg.

## upgrading thorvg

- decide the commit in the https://github.com/thorvg/thorvg repo
  to upgrade to
- update `thorvg-commit-hash` with the new commit hash,
  and commit and push the change
- wait for the github workflow actions to complete
- run `download.go` in the local clone of this repo
  - 2 arguments are required
    1. a github api token, such as a PAT
    1. the path to a local clone of the https://github.com/pekim/thorvg repo
  - for example if the `thorvg` repo is in a sibling directory,
    `go run artifacts/download.go $GITHUB_PAT ../thorvg`

Following the above, the local `thorvg` repo should have updated shared libraries
(`libthorvg_<goos>_<goarch>.go`),
and possibly an updated `thorvg_capi.h`.
If there are api changes in `thorvg_capi.h` then they will need to be accomodated
before committing.
