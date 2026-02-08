import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@utils": path.resolve(__dirname, "./src/utils"),
      "@services": path.resolve(__dirname, "./src/services"),
      "@hooks": path.resolve(__dirname, "./src/hooks"),
      "@store": path.resolve(__dirname, "./src/store"),
      "@svg" : path.resolve(__dirname, "./src/svg"),
      "@libs" : path.resolve(__dirname, "./src/libs"),
      "@pages" : path.resolve(__dirname, "./src/pages"),
      "@layouts" : path.resolve(__dirname, "./src/layouts"),
      "@types" : path.resolve(__dirname, "./src/types"),
      "@schemas" : path.resolve(__dirname, "./src/schemas"),
    },
  },
})
