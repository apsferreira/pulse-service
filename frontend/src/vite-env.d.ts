/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_SERVICE_TOKEN: string
  readonly VITE_AUTH_SERVICE_URL: string
  readonly VITE_AUTH_MODE: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
