<!-- HEADER -->
<br />
<p align="center">
  <a href="https://github.com/smilyorg/photofield">
    <img src="assets/wide-logo.png" alt="Timeship">
  </a>

  <h3 align="center">Timeship</h3>

  <p align="center">
    Your files. Any time.
    <br />
    <br />
  </p>
</p>



<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about">About</a>
    </li>
    <li><a href="#features">Features</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
    </li>
    <li>
      <a href="#development">Development</a>
    </li>
    <li>
      <a href="#configuration">Configuration</a>
    </li>
    <li><a href="#built-with">Built With</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

## About

**Browse and restore files from filesystem snapshots.**

Timeship is a self-hosted web-based snapshot browser that allows you to explore and navigate through ZFS snapshots with an intuitive interface. It provides a seamless way to view historical versions of your files and directories across different points in time.

## Features

* **ZFS snapshot support**. Native support for ZFS snapshots with automatic timestamp parsing from common snapshot naming patterns.
* **File system navigation**. Browse files and directories within snapshots just like a regular file browser.
* **Single binary**. Self-contained Go binary with embedded UI, easy to deploy.
* **Read-only access**. Never modifies your snapshots or filesystem - completely safe to use.

## Getting Started

### Prerequisites

* [Go] 1.25+ - for building the API server
* [Node.js] - for building the frontend
* [Task] - for running build commands

### Installation

1. Clone the repository
   ```sh
   git clone https://github.com/SmilyOrg/timeship.git
   cd timeship
   ```

2. Build the project
   ```sh
   task build
   ```

3. Run the server
   ```sh
   cd api
   ./api
   ```

4. Open http://localhost:8080 in your browser

### Docker

You can also run Timeship using Docker:

1. Build the Docker image
   ```sh
   task docker:build
   ```
   
   Or manually:
   ```sh
   docker build -t timeship .
   ```

2. Run the container
   ```sh
   docker run -p 8080:8080 -v /your/zfs/pool:/data:ro timeship
   ```

   Or use docker-compose:
   ```sh
   docker compose up -d
   ```

3. Open http://localhost:8080 in your browser

**Note:** Make sure to mount your ZFS datasets or snapshot directories as volumes when running the container.

### Environment Variables

* `TIMESHIP_ROOT` - Root directory to serve (defaults to current working directory)

## Development

### Running in Development Mode

Run the API server with auto-reload:
```sh
task api
```

Run the UI development server:
```sh
task ui
```

The UI dev server will run on http://localhost:5173 and proxy API requests to the API server on port 8080.

## Configuration

### ZFS Snapshot Patterns

Timeship automatically detects and parses common ZFS snapshot naming patterns:
- `auto-weekly-2025-11-09_00-00`
- `auto-hourly-2025-11-09_13-30`
- `backup-2025-11-09_14-30-45`
- `snapshot_20251109_143045`
- `daily-2025-11-09`

## Built With

* [Go] - Backend API server
* [Vue 3] - Frontend framework
* [OpenAPI] - API specification and code generation
* [Task] - Build tool

## Roadmap

- [x] File restoration functionality
- [x] Pre-built binaries for releases
- [x] Docker container support
- [x] Text file preview
- [x] Docker container support
- [ ] Image file preview
- [ ] Configuration file support (YAML/JSON)
- [ ] Configurable snapshot name patterns via config file
- [ ] Authentication and authorization
- [ ] Mobile-responsive design
- [ ] Keyboard shortcuts
- [ ] Snapshot source: git commits
- [ ] Snapshot source: borg backups
- [ ] Diff view between snapshots
- [ ] Search within snapshots
- [ ] Timeline visualization
- [ ] File metadata display
- [ ] Snapshot comparison view
- [ ] Dark mode
- [ ] Caching layer for faster browsing
- [ ] Support for other snapshot systems (btrfs, LVM)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to
discuss what you would like to change.

## License

Distributed under the MIT License. See `LICENSE` for more information.

[Go]: https://golang.org/
[Node.js]: https://nodejs.org/
[Vue 3]: https://v3.vuejs.org/
[Task]: https://taskfile.dev/
[OpenAPI]: https://www.openapis.org/
[Photofield]: https://github.com/SmilyOrg/photofield
[oapi-codegen]: https://github.com/oapi-codegen/oapi-codegen
