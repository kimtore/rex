CREATE TABLE library
(
    id                INTEGER primary key AUTOINCREMENT,
    artist            varchar(64),
    title             varchar(64),
    album             varchar(64),
    year              varchar(16),
    genre             varchar(64),
    tracknumber       varchar(3),
    location          integer REFERENCES track_locations (location),
    comment           varchar(256),
    url               varchar(256),
    duration          float,
    bitrate           integer,
    samplerate        integer,
    -- https://github.com/ambientsound/rex/issues/3
    --cuepoint          integer,
    bpm               float,
    wavesummaryhex    blob,
    channels          integer,
    datetime_added    text,
    mixxx_deleted     integer,
    played            integer,
    header_parsed     integer,
    filetype          varchar(8),
    replaygain        float,
    timesplayed       integer,
    rating            integer,
    key               varchar(8),
    beats             BLOB,
    beats_version     TEXT,
    composer          varchar(64),
    bpm_lock          INTEGER,
    beats_sub_version TEXT,
    keys              BLOB,
    keys_version      TEXT,
    keys_sub_version  TEXT,
    key_id            INTEGER,
    grouping          TEXT,
    album_artist      TEXT,
    coverart_source   INTEGER,
    coverart_type     INTEGER,
    coverart_location TEXT,
    coverart_hash     INTEGER,
    replaygain_peak   REAL,
    tracktotal        TEXT,
    color             INTEGER
);
CREATE TABLE track_locations
(
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    location           varchar(512) UNIQUE,
    filename           varchar(512),
    directory          varchar(512),
    filesize           INTEGER,
    fs_deleted         INTEGER,
    needs_verification INTEGER
);
CREATE TABLE crates
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          varchar(48) UNIQUE NOT NULL,
    count         INTEGER DEFAULT 0,
    show          INTEGER DEFAULT 1,
    locked        INTEGER DEFAULT 0,
    autodj_source INTEGER DEFAULT 0
);
CREATE TABLE crate_tracks
(
    crate_id INTEGER NOT NULL REFERENCES crates (id),
    track_id INTEGER NOT NULL REFERENCES "library" (id),
    UNIQUE (crate_id, track_id)
);
CREATE TABLE Playlists
(
    id            INTEGER PRIMARY KEY,
    name          varchar(48),
    position      INTEGER,
    hidden        INTEGER DEFAULT 0 NOT NULL,
    date_created  datetime,
    date_modified datetime,
    locked        INTEGER DEFAULT 0
);
CREATE TABLE PlaylistTracks
(
    id                INTEGER PRIMARY KEY,
    playlist_id       INTEGER REFERENCES Playlists (id),
    track_id          INTEGER REFERENCES "library" (id),
    position          INTEGER,
    pl_datetime_added text
);
