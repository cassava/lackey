lackey
=======

The primary purpose that **lackey** is to manage a lower-quality version of your
music library. This comes in handy when you want to compress your collection of
FLACs to something you can put on your phone (for example).

The lackey program is distributed under the [MIT License](LICENSE).

### Installation

If you have [Go](https://golang.org) installed, installing lackey is easy:

```sh
go get -u github.com/cassava/lackey
cd $GOPATH/src/github.com/cassava/lackey
go install ./cmd/...
```

### Usage
At the moment, lackey's functionality is quite basic. It comes with several
commands and flags that you can use. There are four commands:

 - **sync** – synchronizes from your high-quality music library
   to a lower-quality mirror
 - **stats** – reads a library and prints information about it
   (this may not be particularly useful for you, but it is for me as the dev)
 - **version** – shows the version and compilation date of lackey
 - **help** – shows the help for lackey and the other commands

Let's say your music library is at `~/music`, and you would like a mirror at
`~/music2go`. Then we would use lackey like this:

```sh
lackey sync --library=~/music --delete-before ~/music2go
```

This will by default use the following options (as shown by `lackey help sync`):

 - it will follow symlinks (`--follow-symlinks=true`)
 - it will unconditionally convert non-MP3 music with the LAME encoder
   with a quality setting of 4 (`--quality=4`). See the LAME encoder on
   what the quality setting means. Lower is better.
 - it will convert existing MP3s if they have a bitrate higher than 256kbps,
   and copy them otherwise (`--threshold=256`)
 - it will copy all data files that are not music
 - it will delete all unexpected files in the destination (like rsync does it,
   essentially)
 - it will use the number of cores as the number of workers to use
   (`--concurrent=4`)

Note that reading in the library can take a few minutes (I hope to improve this
in the future); it's insignificant compared to converting and copying files, but
you probably don't want to do this when there's nothing to do.

With these settings, I can reduce a 110GB library to about 30GB. If you want it
to take up even less space, you can increase the quality setting and reduce the
threshold at which it is converted.
If you want to configure even more, let me know and I'll see what I can do.

Enjoy!
