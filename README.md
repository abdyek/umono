# Umono
**Umono** is the vendor lock-in killer CMS.

## v0.5.0 Notes
v0.5.0 introduces breaking changes (language, admin UI, tooling).  
Please read the migration guide before upgrading:

👉 https://umono.io/migration-v0-5

## Usage
### Easy Way (Recommended)
Use the official CLI to manage your Umono websites easily:
👉 [Umono CLI](https://github.com/umono-cms/cli)

### Manually
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
