# Umono
**Umono** is the vendor lock-in killer CMS.

## Install
```sh
curl -fsSL https://umono.io/install.sh | sh
```

## Usage
Use the CLI to create a new project and Up

```
umono create my-website
cd my-website
umono up
```

👉 http://127.0.0.1:8999/admin

## Development
### TailwindCSS Installation
TailwindCSS is used for the admin UI and must be installed locally.
```bash
mkdir -p bin
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.18/tailwindcss-linux-x64
mv tailwindcss-linux-x64 bin/tailwindcss
chmod +x bin/tailwindcss
```

### Live Reload
Umono uses **air** for live reload.

Runs on **port 9000**:
```
air
```

### Contributing
Umono is a small, opinionated project built on the idea of simplicity and clear separation of concerns. I genuinely appreciate any help—whether it's fixing a simple typo, improving the documentation, or tackling a bug.

If you are looking for a place to start, feel free to check out the issues labeled `good first issue`. For larger changes or new features, please open an issue first so we can chat about it before you write any code.

Please take a look at our [CONTRIBUTING.md](CONTRIBUTING.md) for a quick overview of the project's philosophy and workflow.
