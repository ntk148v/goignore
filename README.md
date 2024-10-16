# GoIgnore

- [GoIgnore](#goignore)
  - [1. Features](#1-features)
  - [2. Install](#2-install)
  - [3. Usage](#3-usage)

> [!WARNING]
> I have switched to a shell script based on [fzf](https://github.com/junegunn/fzf) to replace goignore. Therefore, this tool won't be updated.

```shell
function gi() {
    selections=$(curl -sL https://www.toptal.com/developers/gitignore/api/list\?format\=lines | fzf -m --height=80% \
        --prompt='▶ ' --pointer='→' \
        --border=sharp \
        --preview='curl -sL https://www.toptal.com/developers/gitignore/api/{}' \
        --preview-window='45%,border-sharp' \
        --prompt='gitignore ▶ ' \
        --bind='ctrl-r:reload(curl -sL https://www.toptal.com/developers/gitignore/api/list\?format\=lines)' \
        --bind='ctrl-p:toggle-preview' \
        --header '
--------------------------------------------------------------
* Tab/Shift-Tab:       mark multiple items
* ENTER:               append the selected to .gitignore file
* Ctrl-r:              refresh the list
* Ctrl-p:              toggle preview
* Ctrl-q:              exit
* Shift-up/Shift-down: scroll the preview
--------------------------------------------------------------
')

    if [[ ${#selections[@]} == 0 ]]; then
        echo "▶ Nothing selected"
        return 0
    fi

    # allow multi-select
    touch $PWD/.gitignore
    while IFS= read -r s; do
        url="https://www.toptal.com/developers/gitignore/api/$s"
        if grep -q "$url" "$PWD/.gitignore"; then
            echo "▶ $s already added in gitignore, skipping ..."
            continue
        fi
        curl -sL "$url" >>$PWD/.gitignore
        echo "▶ Appended: $s"
    done <<<"$selections"
}
```

A `.gitignore` wizard which generates `.gitignore` files from the command line for you. Inspired by [joe](https://github.com/karan/joe)

## 1. Features

- No installation necessary - just use the binary.
- Works on Linux, Windows & MacOS.
- Interactive user interface with [bubbletea](https://github.com/charmbracelet/bubbletea): Pagination, Filtering, Help...
- Supports all Github-supported [.gitignore files](https://github.com/github/gitignore.git).

## 2. Install

- Download the latest binary from the [Release page](https://github.com/ntk148v/goignore/releases). It's the easiest way to get started with `goignore`.
- Make sure to add the location of the binary to your `$PATH`.

## 3. Usage

- Just run.

```bash
chmod a+x goignore
goignore
```

- At the first time, `goignore` will download the Gitignore templates from Github. It may take a few seconds (depend on your network).

- The list of gitignore templates.

![](./screenshots/start.png)

- Show help.

![](./screenshots/help.png)

- Filter.

![](./screenshots/filter1.png)

![](./screenshots/filter2.png)

- Result, the current gitignore is updated.

![](./screenshots/diff.png)
