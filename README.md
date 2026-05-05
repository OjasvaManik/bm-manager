# bm-manager

A minimal CLI bookmark manager written in Go.

Stores bookmarks in a JSON file, supports aliases, and opens links directly in a browser. No external dependencies.

---

## Features

* Add bookmarks with optional aliases
* List bookmarks in a tree-style view
* Delete bookmarks by index
* Open bookmarks by index or alias
* Configurable default browser
* Optional browser override per command

---

## Installation

### Build

```bash
go build -o bm
```

### Move to PATH

```bash
sudo mv bm /usr/local/bin/
```

Now run:

```bash
bm
```

---

## First Run

On first execution, the following is created:

```
~/.config/bm-manager/bm.conf
```

You **must** edit this file.

Example:

```
location = /home/youruser/bookmarks
browser = firefox
```

* `location` → directory where `bookmarks.json` will be stored
* `browser` → default browser command

---

## Usage

### List bookmarks

```bash
bm
bm list
```

---

### Add bookmark

```bash
bm add "Name" "URL"
bm add "Name" "URL" alias1,alias2
```

---

### Delete bookmark

```bash
bm del 1
```

Index is based on the list output.

---

### Open bookmark

```bash
bm open 1
bm open alias
```

Override browser:

```bash
bm open alias -b zen-browser
```

---

## Data Format

Bookmarks are stored as JSON:

```json
[
  {
    "name": "Google",
    "url": "https://google.com",
    "alias": ["g"]
  }
]
```

Aliases are optional.

---

## How It Works

* Reads config from `~/.config/bm-manager/bm.conf`
* Loads bookmarks from `bookmarks.json`
* Commands mutate in-memory slice
* Writes back using formatted JSON
* Opens URLs using `exec.Command`

---

## Notes

* Indexes are 1-based
* Alias matching is exact
* No fuzzy matching or UI layer
* Errors are fatal and explicit

---

## Summary

A simple CLI tool replacing shell scripts and fuzzy launchers with a structured, predictable implementation in Go.

No abstractions beyond what is required.

