![Example Tests](https://github.com/caesurus/riptracer/actions/workflows/examples.yaml/badge.svg)
![Unit Tests](https://github.com/caesurus/riptracer/actions/workflows/unittests.yaml/badge.svg)


# riptracer

Execution tracer written in `go`. Think strace/ltrace for arbitrary code locations. Set breakpoints, manipulate memory/registers, etc...

## Why?

Do you know how implement a software breakpoint, or a how to set a hardware breakpoint? I had used both in `gdb` for years without understanding exactly how these are implemented. I wanted to change that, and so what better way than implementing a debugger yourself?

I've been a longtime fan of [`usercorn`](https://github.com/lunixbochs/usercorn). I even have a repo of example script [`usercorn_examples`](https://github.com/Caesurus/usercorn_examples). But there are some drawbacks to having to emulate everything. Not all the system calls are implemented, and if the binary does threading, we're probably in for a rough time. I wanted to debug a threaded binary without emulating, and have custom debug functionality.

## Why not just use `gdb`?

With the power of `gdb` and some `gdb scripts` we'd be able to do similar functionality, but we'd need `gdb` on our target system, along with python for the scripting etc... All of which is fine, and totally possible, but a nicely compiled `go` binary can be deployed without having to worry about the dependencies needed. I want to spend time learning and debugging, not cross-compiling and in dependency hell. 

## Disclaimer

This is a toy project, I'll update it for as long as I find it useful and interesting. My aim here is not to rewrite `strace/strace/dtrace/gdb/rr` etc... You should totally use those for anything serious. 
