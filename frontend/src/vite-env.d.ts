/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_PUBLIC_BUCKET_URL: string;
  readonly FIREBASE_API_KEY: string;
  readonly FIREBASE_AUTH_DOMAIN: string;
  readonly FIREBASE_AUTH_EMULATOR_URL: string?;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
