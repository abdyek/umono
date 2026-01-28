# Umono
**Umono** is the vendor lock-in killer CMS.

## v0.5.x Notes
Although v0.5 is a minor release, it introduces significant structural changes. With this release:

- The **admin-ui** repository has been deprecated. The Admin UI is now an internal part of Umono.
- The **build** repository has been deprecated. Umono can now exist as a single binary.
- [Compono](https://github.com/umono-cms/compono) has replaced **UmonoLang**. Compono was designed specifically for Umono to produce more predictable outputs.
- **A default frontend** has been introduced. Until the theme system arrives in a future release, this aims to provide a more polished default look for Umono.

You can create a fresh Umono v0.5.x instance and migrate by copying `umono.db` and `.env`. If you need help please open an issue.

## Usage
### Easy Way (Coming Soon - WIP)
Use the official CLI to manage your Umono websites easily:

👉 [Umono CLI](https://github.com/umono-cms/cli)

### Manually
Dependencies
- gcc/clang (for CGO)
- libc / glibc development headers
- SQLite development library
- Go (CGO enabled)

Clone an empty Umono project:
```bash
git clone https://github.com/umono-cms/umono my-website
cd my-website
```
Create .env file:
```bash
cp .env.example .env
```
**Don’t forget to update the .env file.**

Build and run:
```bash
go build -o umono ./cmd/umono
./umono
```
Runs on **port 8999**.

✨ You are ready to create your first page

👉 http://127.0.0.1:8999/admin

## Development
### Live Reload
Umono uses **air** for live reload.

Runs on **port 9000**:
```
air
```
### TailwindCSS Installation
TailwindCSS is used for the admin UI and must be installed locally.
```bash
mkdir -p bin
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.18/tailwindcss-linux-x64
mv tailwindcss-linux-x64 bin/tailwindcss
chmod +x bin/tailwindcss
```
Watch:
```bash
./bin/tailwindcss -i assets/input.css -o public/css/style.css --watch
```
