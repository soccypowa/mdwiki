# mdwiki

Serves a folder of interlinked markdown files (a markdown "wiki repo", with
an `index.md` per folder) as a local website. Renders markdown server-side
with [goldmark](https://github.com/yuin/goldmark), rewrites relative `.md`
links to server routes, and uses htmx's `hx-boost` for fast, full-page-reload-free
navigation — no JavaScript build step, no client-side router.

Also supplies a simple search function in the served md files.

## Usage

```
mdwiki <folder> <port>
mdwiki -dir <folder> -port <port>
```

- `<folder>` — root of your markdown wiki (the folder containing the
  top-level `index.md`)
- `<port>` — defaults to 8888 if omitted

Visiting `/` serves `<folder>/index.md`. Visiting `/foo` serves
`<folder>/foo.md` if it exists, otherwise `<folder>/foo/index.md`. Any
non-markdown file (images, PDFs, etc.) referenced from your docs is served
as a static file straight from disk.

## How link rewriting works

Your links stay exactly as they are in the source files, e.g.:

```markdown
See the [setup guide](./setup.md) or go [back](../index.md).
```

At render time, `mdwiki` walks the parsed markdown AST and strips the
`.md`/`.markdown` suffix from relative link destinations only — absolute
URLs, `#anchors`, and `mailto:` links are left untouched. That link becomes
`./setup` and `../index`, which line up with the routes above. You don't
need to change anything in your existing files.

## Building from source

Requires Go 1.22+.

```bash
make build
```


## Running as a background service

### Windows (services)

1. Download a servicewrappe, for example:
   
   https://github.com/soccypowa/win-svcwrapper

2. Follow the instructions to create a service on your windows box.

### Linux (systemd)

1. Copy the binary somewhere permanent
   ```sh
   sudo cp mdwiki /usr/local/bin/mdwiki
    sudo chmod +x /usr/local/bin/mdwiki
   ```
2. Create and edit `/etc/systemd/system/mdwiki.service`
   ```ini
   # /etc/systemd/system/mdwiki.service
   [Unit]
   Description=mdwiki - local markdown wiki server
   After=network.target
   
   [Service]
   Type=simple
   ExecStart=/usr/local/bin/mdwiki /path/to/your/docs 8888
   Restart=on-failure
   RestartSec=2
   
   # Run as your normal user, not root — mdwiki only needs read access to the docs folder
   User=YOURUSERNAME
   Group=YOURUSERNAME
   
   # Light sandboxing, safe defaults for a simple read-only local server
   NoNewPrivileges=true
   ProtectSystem=strict
   ProtectHome=read-only
   PrivateTmp=true
   
   [Install]
   WantedBy=multi-user.target
   ```

3. Enable and start the service
   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable --now mdwiki
   ```

### macOS (launchd)

1. Copy the binary somewhere permanent, e.g.:
   ```bash
   sudo cp mdwiki /usr/local/bin/mdwiki
   sudo chmod +x /usr/local/bin/mdwiki
   xattr -d com.apple.quarantine /usr/local/bin/mdwiki
   ```
2. Create annd edit `com.local.mdwiki.plist`
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
       <dict>
           <key>Label</key>
           <string>com.local.mdwiki</string>   
           <key>ProgramArguments</key>
           <array>
               <!-- binary absolute path -->
               <string>/usr/local/bin/mdwiki</string>
               <!-- absolute path do documentation folder -->
               <string>/path/to/your/docs</string>
               <!-- port to serve on -->
               <string>8888</string>
           </array>   
           <key>RunAtLoad</key>
           <true />   
           <key>KeepAlive</key>
           <true />   
           <key>StandardOutPath</key>
           <string>/tmp/mdwiki.log</string>
           <key>StandardErrorPath</key>
           <string>/tmp/mdwiki.log</string>
       </dict>
   </plist>
   ```
3. Install and start it:
   ```bash
   cp com.local.mdwiki.plist ~/Library/LaunchAgents/
   launchctl load ~/Library/LaunchAgents/com.local.mdwiki.plist
   ```
   It will now start automatically at login and restart if it crashes
   (`KeepAlive`). Logs go to `/tmp/mdwiki.log`.

   To stop/uninstall:
    ```bash
    launchctl unload ~/Library/LaunchAgents/com.local.mdwiki.plist
    rm ~/Library/LaunchAgents/com.local.mdwiki.plist
    ```

## Notes

- htmx is vendored into the binary via `go:embed` (`static/htmx.min.js`),
  so the whole thing works fully offline — no CDN dependency at runtime.
- Path resolution is safe against `../` traversal outside the docs root
  (Go's `path.Clean` on a rooted path can't escape above `/`).
- There's no live-reload / file-watcher — each request re-reads and
  re-renders the relevant `.md` file from disk, so edits show up on the
  next page load with no restart needed.
