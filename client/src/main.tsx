import { StrictMode, useEffect } from "react";
import { Toaster } from "sonner";
import { BrowserRouter } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { createRoot } from "react-dom/client";
import { queryClient } from "@/libs/react-query";
import { useAuthStore } from "@/store/auth.store"; 
import "./index.css";
import App from "./App.tsx";

// 🔥 Theme Manager Component
const ThemeManager = ({ children }: { children: React.ReactNode }) => {
  const theme = useAuthStore((state) => state.user?.preferences?.theme || "system");

  useEffect(() => {
    const root = window.document.documentElement;

    const applyTheme = () => {
      root.classList.remove("light", "dark");

      if (theme === "system") {
        const systemPrefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
        root.classList.add(systemPrefersDark ? "dark" : "light");
      } else {
        root.classList.add(theme);
      }
    };

    applyTheme();

    if (theme === "system") {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      const handleChange = () => applyTheme();
      mediaQuery.addEventListener("change", handleChange);
      return () => mediaQuery.removeEventListener("change", handleChange);
    }
  }, [theme]);

  return (
    <>
      {children}
      <Toaster 
        richColors 
        position="top-right" 
        theme={theme as "light" | "dark" | "system"} 
      />
    </>
  );
};

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ThemeManager>
          <App />
        </ThemeManager>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>
);