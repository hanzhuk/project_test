{
  "name": "{{.ProjectName}}-web",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "gen:api": "openapi-typescript http://localhost:8080/openapi.json -o src/api/schema.d.ts"
  },
  "dependencies": {
    "svelte": "^5.0.0",
    "openapi-fetch": "^0.13.0"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "typescript": "^5.6.0",
    "vite": "^6.0.0",
    "openapi-typescript": "^7.4.0",
    "tailwindcss": "^3.4.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0"
  }
}
