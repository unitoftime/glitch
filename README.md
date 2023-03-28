`Warning: This library is very much a work in progress. But you're welcome to check it out and provide feedback/bugs if you want.`

## Overview
Glitch is a shader based rendering library built on top of OpenGL and WebGL. At a high level, I'd like glitch to be data driven. Shaders are just programs that run on the GPU, so my objective is to make Glitch a platform that makes it easier to do the things that are hard in rendering:
1. Efficiently ordering, moving, and batching data to the GPU
2. Efficiently executing programs on that copied data

## Platform Support
Currently, we compile to:
 * Desktop (Windows, Linux)
 * Browser (via WebAssembly)

Platforms that I'd like to add, but haven't added or haven't tested:
 * Desktop (MacOS - OpenGL is deprecated, I also don't own a mac. So it's hard to test)
 * Mobile Apps
 * Mobile Browsers

## Usage
You can look at the examples folder, sometimes they go out of date, but I try to keep them working. Because APIs are shifting I don't have definite APIs defined yet.
