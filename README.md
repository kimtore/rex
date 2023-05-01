# rex: rekordbox exporter for Mixxx

Open source mixing or library software should be able to create Rekordbox
compatible export files, so that they can be played on Pioneer equipment in
venues all over the world.

This project is my attempt at getting closer to this goal, and leans heavily on the work done by others.
A good starting point is: https://djl-analysis.deepsymmetry.org/rekordbox-export-analysis/

## Project state

Do not use files generated from this project on a live gig, it probably won't work and you'll be miserable.
That said, it is possible to create PDB files that can be opened in Rekordbox.
These files have also been tested on a few Pioneer devices and are usable to varying degrees.
Trying to import them on a Denon Prime 4 results in something happening, but no library.

I figured out some more fields from various tables, and also a bit how the table structure should be built up.
The important stuff is in the [rekordbox package](pkg/rekordbox) subdirectories.
Especially the stuff in [dbengine](pkg/rekordbox/dbengine) and [page](pkg/rekordbox/page)
might be of particular interest. Many tests are broken, they might not be relevant.

## Prerequisites

This software exports Mixxx libraries, you must have a fairly recent version of Mixxx installed.

REX has only been tested on Arch Linux with Mixxx 2.3.4.
Your results may vary.

REX requires FFMPEG to transcode audio files.

## Generate exports

This software has been tested successfully with Go 1.20.

Use [REX](cmd/rex/main.go) to generate PDB files from your Mixxx library.

```
go build -o rex cmd/rex/main.go
./rex -root /path/to/USB
```

Your copied audio files will be put in the `rex` folder on the USB media,
you can change this with `-trackdir my-audio-files`. If your Mixxx library
is not in the correct place, you can change it with `-mixxxdb /path/to/mixxxdb.sqlite3`.

These features are NOT supported yet:

* Waveforms
* Beat grid
* Hot Cue

## Export file analysis

Use [Analyze](cmd/analyze/main.go) to introspect what's going on inside the files:

```
go build -o analyze cmd/analyze/main.go
./analyze -index -rows /path/to/USB/PIONEER/rekordbox/export.pdb
```
