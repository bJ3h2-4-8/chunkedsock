# chunkedsock

opens a listen socket to connect "chunked" (or limited) to a remote port, effectively reopening the connection after a specified amount of kBs, but behaving transparent for the incoming connection.

## Why?
this was mainly a workaround, overcoming a freezing socket on an embedded device which operates on the next turn as-nothing-ever-ever-happened™ ...so maybe it might be useful in similar situations.

## How?

clone or download the repo as usual and build it, like so:

``go build -o  chunkedsock cmd/chunkedsock/main.go``

afterwards, one might call it without arguments for help.

## Example!
Here is a simple invocation if you'd like to bridge a buggy connection of the target address _192.168.55.8_ at port _8553_ and have it listen on your host machine on port _9000_:

``chunkedsock -l :9000 -t 192.168.55.8:8553``

this creates a _chunked_ connection of 200kb each, which is the default if you leave out the _-c_ or _--chunksize_ option.


## More?!
Yeah sure, it even tries to reduce the latency which would occure by opening the next connection in advance, so it can reduce this effect pretty easy, and even if this does not work - it will try again after the active chunked transmission has ended.




