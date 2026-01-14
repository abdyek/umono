# Umono
Umono is a content management system written golang.

> **🚧 v0.5 Development in Progress**
> 
> Major changes coming in v0.5:
> - ✨ HTMX integration (replacing Vue.js admin)
> - ✨ Compono language support (replacing UmonoLang)
> - ✨ Improved architecture
> 
> **Current stable version:** v0.4.x (use `main` branch)  
> **Development version:** v0.5-dev (this branch)

## Demo
You can demo it on [build](https://github.com/umono-cms/build?tab=readme-ov-file#demo) repository.

## Production
You can use it on [build](https://github.com/umono-cms/build?tab=readme-ov-file#production) repository.

## Development

### Requirements
- Golang
- Node.js

### Admin UI
Umono has a admin UI written Vue.
#### Clone
```
git clone https://github.com/umono-cms/admin-ui
```

#### Change directory
```
cd admin-ui
```

#### Install packages
```
npm install
```

#### Start process for tailwindcss
```
npx tailwindcss -i ./input.css -o ./src/style.css --watch
```

#### Run
```
npm run dev
```

### Backend
#### Clone
```
git clone https://github.com/umono-cms/umono
```

#### Change directory
```
cd umono
```

#### .env file
Copy .env file from .env-example and edit it
```
cp .env-example .env
```

#### Run server
```
go run .
```
