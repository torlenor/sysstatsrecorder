# System Stats Recorder for (only Windows for now)

This simple tool records current CPU and memory usage to a csv file.

## Requirements

Go >= 1.12

Please check out https://golang.org/doc/install on how to install Go (it is easy, no worries).

## Building System Stats Recorder

### Windows

To build this project type

```bash
build.bat
```

in the top directory of this repositoty.

### Crosscompilation on Linux

WIP

### Running System Stats Recorder

```bash
bin\gocommrecorder.exe -f outputFileNamePrefix -t 1000
```

where outputFileNamePrefix is the prefix used for the output file (timestamp and .csv will be appended automatically) and 1000 is the interval to use for recording in ms.
