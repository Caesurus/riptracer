![Example Tests](https://github.com/caesurus/riptracer/actions/workflows/examples.yaml/badge.svg)
![Unit Tests](https://github.com/caesurus/riptracer/actions/workflows/unittests.yaml/badge.svg)


# riptracer

Execution tracer written in `go`. Think strace/ltrace for arbitrary code locations. Set breakpoints, manipulate memory/registers, etc...

## Why?

I created this because I wanted to learn more about implementing a debugger on Linux. 

I've been a longtime fan of [`usercorn`](https://github.com/lunixbochs/usercorn). I even have a repo of example script [`usercorn_examples`](https://github.com/Caesurus/usercorn_examples). But there are some drawbacks to having to emulate everything. Not all the system calls are implemented, and if the binary does threading, we're probably in for a rough time. I wanted to debug a threaded binary without emulating, and have custom debug functionality

## Why not just use `gdb`?

With the power of `gdb` and some `gdb scripts` we'd be able to do similar functionality, but we'd need `gdb` on our target system, along with python for the scripting etc... All of which is fine, and totally possible, but a nicely compiled `go` binary can be deployed without having to worry about the dependencies needed. I want to spend time learning and debugging, not cross-compiling and in dependency hell. The real reason is what I already gave, I wanted to know how debuggers work in linux. What better way to implement one yourself.

## Disclaimer

This is a toy project, I'll update it for as long as I find it useful and interesting. My aim here is not to rewrite `strace/strace/dtrace/gdb/rr` etc... You should totally use those for anything serious. 

