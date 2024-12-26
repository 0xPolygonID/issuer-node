import react from "@vitejs/plugin-react";
import * as dotenv from "dotenv";
import { defineConfig } from "vite";
import { checker } from "vite-plugin-checker";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

dotenv.config();

// eslint-disable-next-line import/no-default-export
export default defineConfig({
  base: process.env.VITE_BASE_URL || "/",
  plugins: [
    checker({
      eslint: {
        lintCommand: 'eslint "./src/**/*.{ts,tsx}"',
      },
      overlay: false,
      typescript: true,
    }),
    react(),
    svgr(),
    tsconfigPaths(),
  ],
});
