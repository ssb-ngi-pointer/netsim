# Network Simulator 
_a simulator for testing [secure scuttlebutt](https://ssb.nz) implementations against each other_

## Goals
The network simulator should..
* be a tool to measure performance metrics before & after partial replication
* be reusable by other scuttlebutts for verifying changes & debugging situations—without requiring any build step to run the tool
* be flexible enough to add new types of peers (e.g. rust)
* provide assurance + insurance that the bedrock of scuttlebutt works as intended

## Usage
```sh
netsim generate <ssb-fixtures-output> # processes ssb-fixtures to a form that netsim consumes
netsim run path-to-sbot1 path-to-sbot2 ... path-to-sbotn
``` 

The `netsim` utility has two commands: 
* `netsim generate` consumes output generated by
  [`ssb-fixtures`](https://github.com/ssb-ngi-pointer/ssb-fixtures) and outputs a _netsim-adapted_
  ssb-fixtures folder, and a netsim test file
* `netsim run` runs the specified netsim test file using the specified sbot implementations

When auto-generating a netsim test, netsim makes use of pre-generated [ssb-fixtures](https://github.com/ssb-ngi-pointer/ssb-fixtures)
to..
* source identities and their public keypair, 
* determine the follow graph, 
* the amount of peers in the simulation, and 
* the total amount of messages

For more options:
```sh
netsim generate -h
netsim run -h
``` 

For more on authoring netsim commands: 
* [`commands.md`](./commands.md) 
* [`ssb-netsim`](https://github.com/ssb-ngi-pointer/ssb-netsim), the nodejs helper library 

### Downloading
To get started quickly, [download a netsim release](https://github.com/ssb-ngi-pointer/netsim/releases).

## Example
Say we want to test an sbot implementation in `~/code/ssb-server`, just to make sure the basics
are still working.

First, we write a netsim test called `basic-test.txt`:
```
# booting
enter peer
hops peer 1
enter server
hops server 1

start peer ssb-server
start server ssb-server
post server
post server
follow peer server
follow server peer
connect peer server
waituntil peer server@latest
stop peer
```

Now, let's run it:
```sh
netsim run --spec basic-test.txt ~/code/ssb-server
```

**Note:** the folder containing the sbot implementation, `ssb-server`, and the `start`
command's last operand are the same. If you want to run more sbots in the same test, just add
them onto the `netsim run` invocation, while making sure to match the folder name with
`start`'s operand.

### Building
If you want to build the code yourself: 

```sh
git clone git@github.com:ssb-ngi-pointer/netsim
cd netsim/cmd/netsim
go build
./netsim
```

For the unbundled commands, see the
[`cmd/`](https://github.com/ssb-ngi-pointer/netsim/tree/main/cmd) folder.

## Simulation Shims
Before you run netsim against an sbot implementation, make sure you have a `sim-shim.sh` script
in the root of the implementation. Simulation shims encapsulate implementation-specific details &
procedures such as: ingesting a [`log.offset`](https://github.com/flumedb/flumelog-offset)
file, passing `hops` and `caps` settings to the underlying sbot, and other details.

The `sim-shim.sh` script is passed, and should use, the following arguments and environment variables:

```sh
DIR="$1"
PORT="$2"
# the following env variables are always set from netsim:
#   ${CAPS}   the capability key / app key / shs key
#   ${HOPS}   a integer determining the hops setting for this ssb node
# if ssb-fixtures are provided, the following variables are also set:
#   ${LOG_OFFSET}  the location of the log.offset file to be used
#   ${SECRET}      the location of the secret file which should be copied to the new ssb-dir
``` 

For go and nodejs examples of sim-shims, see [`sim-shims/`](./sim-shims).

**Note:** the file must be named `sim-shim.sh` for the netsim to work.

## Required muxrpc calls
In order to test different implementations against each other, netsim makes heavy use of
Secure Scuttlebutt [`muxrpc`](https://github.com/ssb-js/muxrpc) calls. For a brief primer, [see
the protocol guide](https://ssbc.github.io/scuttlebutt-protocol-guide/#rpc-protocol).

Currently, the following calls are required to be implemented before an implementation is
testable in netsim:

**Essential**
* `createHistoryStream`
* `whoami`
* `publish`
* `replicate.upto`
* `conn.connect`
* `conn.disconnect`

**Extras**
* `createLogStream` used by the `log` command
* `friends.isFollowing` used by `isfollowin` / `isnotfollowing`
