/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_PULSE_API_URL: string
  readonly VITE_SERVICE_TOKEN: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
