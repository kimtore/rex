# rex: rekordbox exporter

Open source mixing or library software should be able to create Rekordbox
compatible export files, so that they can be played on Pioneer equipment in
venues all over the world.

This project is my attempt at getting closer to this goal, and leans heavily on the work done by others.
A good starting point is: https://djl-analysis.deepsymmetry.org/rekordbox-export-analysis/

## Project state

It is possible to create PDB files that can be opened in Rekordbox.
I haven't been able to test them on real Pioneer hardware yet.
Trying to import them on a Denon Prime 4 results in something happening, but no library.

Do not use files generated from this project on a live gig, it probably won't work and you'll be miserable.

I figured out some more fields from various tables, and also a bit how the table structure should be built up.

The important stuff is in the [rekordbox package](pkg/rekordbox) subdirectories.
Especially the stuff in [dbengine](pkg/rekordbox/dbengine) and [page](pkg/rekordbox/page)
might be of particular interest.

Many tests are broken, they might not be relevant.

## Tools and development

This software has been tested successfully with Go 1.20.

Use [REX](cmd/rex/main.go) to generate PDB files:
```
go build -o rex cmd/rex/main.go
./rex -root /path/to/USB -scan /path/to/USB/mymusic
```

Use [Analyze](cmd/analyze/main.go) to introspect what's going on inside the files:
```
go build -o analyze cmd/analyze/main.go
./analyze -index -rows /path/to/USB/PIONEER/rekordbox/export.pdb
```
