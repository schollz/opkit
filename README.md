# postsolarpunk

this repository utilizes postsolarpunk's generously provided sample library of pulsar-23 sounds and randomly generates patch files that can be uploaded to an OP-Z or OP-1 with correct splice points. patch files are created with different directories (Kick, Snare, Bass, etc.) and they are automatically tunes to different notes and DC offset removed with a highpass filter. other than that - the samples are exactly as I found them.

## usage

prerequisites: make sure that you have Go, sox, and ffmpeg installed.

first download postsolarpunk's kit. then unzip it into this directory.

```
unzip 'pulsar-23 postsolarpunk pack.zip'
```

build the software.

```
go build -v
```

now run the converter:

```
./postsolarpunk --convert
```

this will generate a directory `psp` with all of the converted files.

now you can generate kits by filtering on the folder names. for example, to generate a "Kick" kit tuned to the C:

```
./postsolarpunk --kit "Kick" --tuned 24 --out kick.aif
```

the result is a single file that fits as many samples as it can within the OP-Z/OP-1 limit (<12 seconds);

![example](https://share.schollz.com/1/goofygila/1.png)

the resulting file sorts the samples from longest to shortest. 
