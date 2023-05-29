# Securing APIs with Spire and OpenZiti

This repository leverages two fantastic open source projects (three including golang):

* [SPIFFE/SPIRE](https://github.com/spiffe/spire)
* [OpenZiti](https://github.com/openziti/ziti)

If you find the amazing levels of security and functionality these projects provide,
go give them a hard-earned GitHub star!

Both projects are focused on securing APIs and both go about it in slightly different
ways. This repo will show you how to take an insecure API, exposed via plain HTTP and
secure it with both SPIRE and OpenZiti. There are four main paths to take to run the 
project. You can:

* [Use no security at all](./src/part1_nosecurity)
* [Use only SPIRE for mTLS](./src/part2_spire)
* [Use OpenZiti for end to end encryption](./src/part3_openziti)
* [Use SPIRE and OpenZiti together](./src/part4_spire_and_openziti)

## Running One or More Examples

Regardless of which of the secure examples you run, to run them
you'll need to make sure you have SPIRE setup. By far, the lowest
friction way of doing this is to just run the [provided helper script](./compile-and-run.sh)
in your bash shell. This script will do _a lot_ for you. 

You are highly encouraged to read the script. It shows you exactly 
what commands need to be run and in what order for them to function. The
script serves as a way for you to read and explore each command,
understand what it does what it does, and why.

### Dependencies, Prerequisties, Assumptions

The script has the following dependencies:

* The files will all be saved to `TMP_DIR` which by default is set to `/tmp/dovholuknf/qcon2023`.
* `go` will be needed to build the samples
* `docker` (and the newer `docker compose`)
* `killall` is used to stop any existing servers (in lieu of something more robust like pid tracking)
* it will use `sudo` to delete the folder at `/tmp/dovholuknf/qcon2023` when it runs
* `curl`, `tar`, `sed` are all needed along with other standard commands: `mv`, `export`, `echo`, `sleep`, `cat`, etc.
* `ip` will be used to find eth0's IP. if you don't have an eth0, find eth0 in the script and update it
* you will _need_ to add: `127.0.0.1   ziti-edge-controller ziti-edge-router` to your `/etc/hosts`
  or you'll need to know how to get `ziti-edge-controller` and `ziti-edge-router` as hostnames routable
  into the docker environment that will spin up
* it'll use `sudo` to run your spire agent as root. this is done so that when workloads attest,
  the agent can figure out who is attempting to attest. Obviously, this is not 'a good idea' but
  it's an easy, expedient way of getting the agent the proper permissions


## Cleaning Up

The script has within it all the cleanup steps you need. This will come down to:
* stopping `docker compose`:

      `docker compose -f $TMP_DIR/docker-compose.yml --env-file=$TMP_DIR/.env -p qcon2023 down -v`

* stopping the SPIRE server, agent and oidc-discovery-provider:

      sudo killall spire-server
      sudo killall spire-agent
      sudo killall oidc-discovery-provider

* Removing any related identities from your locally running tunneler (if any)