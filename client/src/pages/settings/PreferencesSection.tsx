import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth.store";
import { useUpdatePreferences } from "@/hooks/useUser";
import SubmitButton from "@/common/SubmitButton";
import { Switch } from "@/common/Switch"; // Aapka naya reusable component
import { Bell, ChevronDown } from "@/svg/Common";
import { DotPattern, Globe, Monitor, Moon, Sun, Theme } from "@/svg/Auth";

export const PreferencesSection = () => {
  const user = useAuthStore((state) => state.user);
  const { updatePreferences, isPending } = useUpdatePreferences();

  const [theme, setTheme] = useState("system");
  const [emailNotifications, setEmailNotifications] = useState(false);
  const [timezone, setTimezone] = useState("UTC");

  useEffect(() => {
    if (user?.preferences) {
      setTheme(user.preferences.theme || "system");
      setEmailNotifications(user.preferences.email_notifications ?? false);
      setTimezone(user.preferences.timezone || "UTC");
    }
  }, [user]);

  const handleSave = () => {
    updatePreferences(
      {
        preferences: {
          theme,
          email_notifications: emailNotifications,
          timezone,
        },
      },
      {
        onSuccess: () => toast.success("Preferences updated successfully!"),
        onError: () => toast.error("Failed to update preferences."),
      },
    );
  };

  if (!user) return null;

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-8 relative overflow-hidden">
      <div className="flex items-start gap-4 mb-6">
        <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
          <Theme className="w-6 h-6 text-zinc-500" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
            Preferences
          </h2>
          <p className="text-sm text-zinc-500">
            Customize your interface and notification settings.
          </p>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80"></div>

      <div className="space-y-8">
        {/* ================= THEME (ENHANCED SVG PATTERNS) ================= */}
        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-5">
            {/* Light Theme Box */}
            <button
              onClick={() => setTheme("light")}
              className={`relative overflow-hidden flex flex-col items-center justify-center p-4 rounded-2xl transition-all duration-300 cursor-pointer ${
                theme === "light"
                  ? "bg-indigo-50/50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300"
                  : "text-zinc-500 dark:text-zinc-40 hover:bg-zinc-50 dark:hover:bg-zinc-800 bg-zinc-50 dark:bg-zinc-800/80"
              }`}
            >
              <DotPattern />

              {/* Light UI Mockup SVG */}
              <svg
                className="w-full h-16 rounded-md mb-3 border border-zinc-200 bg-zinc-50 z-10"
                viewBox="0 0 100 60"
                fill="none"
              >
                <rect x="0" y="0" width="100" height="60" fill="#f4f4f5" />
                <rect
                  x="6"
                  y="6"
                  width="88"
                  height="12"
                  rx="2"
                  fill="#ffffff"
                  stroke="#e4e4e7"
                  strokeWidth="1"
                />
                <rect
                  x="6"
                  y="24"
                  width="55"
                  height="28"
                  rx="2"
                  fill="#ffffff"
                  stroke="#e4e4e7"
                  strokeWidth="1"
                />
                <rect
                  x="67"
                  y="24"
                  width="27"
                  height="10"
                  rx="2"
                  fill="#ffffff"
                  stroke="#e4e4e7"
                  strokeWidth="1"
                />
              </svg>
              <div className="flex items-center gap-1.5 z-10">
                <Sun
                  className={`w-4 h-4 ${theme === "light" ? "text-indigo-600 dark:text-indigo-400" : ""}`}
                />
                <span className="font-semibold text-sm">Light</span>
              </div>
            </button>

            {/* Dark Theme Box */}
            <button
              onClick={() => setTheme("dark")}
              className={`relative overflow-hidden flex flex-col items-center justify-center p-4 rounded-2xl transition-all duration-300 cursor-pointer ${
                theme === "dark"
                  ? "bg-indigo-50/50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300"
                  : "text-zinc-500 dark:text-zinc-40 hover:bg-zinc-50 dark:hover:bg-zinc-800 bg-zinc-50 dark:bg-zinc-800/80"
              }`}
            >
              <DotPattern />
              {/* Dark UI Mockup SVG */}
              <svg
                className="w-full h-16 rounded-md mb-3 shadow-xs border border-zinc-800 bg-zinc-900 z-10"
                viewBox="0 0 100 60"
                fill="none"
              >
                <rect x="0" y="0" width="100" height="60" fill="#18181b" />
                <rect
                  x="6"
                  y="6"
                  width="88"
                  height="12"
                  rx="2"
                  fill="#27272a"
                  stroke="#3f3f46"
                  strokeWidth="1"
                />
                <rect
                  x="6"
                  y="24"
                  width="55"
                  height="28"
                  rx="2"
                  fill="#27272a"
                  stroke="#3f3f46"
                  strokeWidth="1"
                />
                <rect
                  x="67"
                  y="24"
                  width="27"
                  height="10"
                  rx="2"
                  fill="#27272a"
                  stroke="#3f3f46"
                  strokeWidth="1"
                />
              </svg>
              <div className="flex items-center gap-1.5 z-10">
                <Moon
                  className={`w-4 h-4 ${theme === "dark" ? "text-indigo-600 dark:text-indigo-400" : ""}`}
                />
                <span className="font-semibold text-sm">Dark</span>
              </div>
            </button>

            {/* System Theme Box */}
            <button
              onClick={() => setTheme("system")}
              className={`relative overflow-hidden flex flex-col items-center justify-center p-4 rounded-2xl transition-all duration-300 cursor-pointer ${
                theme === "system"
                  ? "bg-indigo-50/50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300"
                  : "text-zinc-500 dark:text-zinc-40 hover:bg-zinc-50 dark:hover:bg-zinc-800 bg-zinc-50 dark:bg-zinc-800/80"
              }`}
            >
              <DotPattern />
              {/* System (Split) UI Mockup SVG */}
              <svg
                className="w-full h-16 rounded-md mb-3 border border-zinc-200 dark:border-zinc-800 z-10 transition-transform"
                viewBox="0 0 100 60"
                fill="none"
              >
                {/* Light Side */}
                <rect x="0" y="0" width="50" height="60" fill="#f4f4f5" />
                <rect
                  x="4"
                  y="6"
                  width="42"
                  height="12"
                  rx="2"
                  fill="#ffffff"
                  stroke="#e4e4e7"
                  strokeWidth="1"
                />
                <rect
                  x="4"
                  y="24"
                  width="42"
                  height="28"
                  rx="2"
                  fill="#ffffff"
                  stroke="#e4e4e7"
                  strokeWidth="1"
                />
                {/* Dark Side */}
                <rect x="50" y="0" width="50" height="60" fill="#18181b" />
                <rect
                  x="54"
                  y="6"
                  width="42"
                  height="12"
                  rx="2"
                  fill="#27272a"
                  stroke="#3f3f46"
                  strokeWidth="1"
                />
                <rect
                  x="54"
                  y="24"
                  width="42"
                  height="28"
                  rx="2"
                  fill="#27272a"
                  stroke="#3f3f46"
                  strokeWidth="1"
                />
              </svg>
              <div className="flex items-center gap-1.5 z-10">
                <Monitor
                  className={`w-4 h-4 ${theme === "system" ? "text-indigo-600 dark:text-indigo-400" : ""}`}
                />
                <span className="font-semibold text-sm">System</span>
              </div>
            </button>
          </div>
        </div>

        {/* ================= NOTIFICATIONS & TIMEZONE ================= */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Email Notifications Toggle (Using Reusable Switch) */}
          <div className="flex items-center justify-between p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30">
            <div className="flex gap-3 items-center">
              <div className="w-10 h-10 rounded-full bg-white dark:bg-zinc-800 flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0">
                <Bell className="w-5 h-5 text-zinc-600 dark:text-zinc-300" />
              </div>
              <div>
                <h4 className="font-medium text-sm text-zinc-900 dark:text-zinc-100">
                  Email Alerts
                </h4>
                <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">
                  Receive updates via email.
                </p>
              </div>
            </div>

            {/* <--- Reusable Component Here ---> */}
            <Switch
              checked={emailNotifications}
              onChange={setEmailNotifications}
            />
          </div>

          {/* Timezone Dropdown */}
          <div className="flex items-center justify-between p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30">
            <div className="flex gap-3 items-center w-full">
              <div className="w-10 h-10 rounded-full bg-white dark:bg-zinc-800 flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0 shadow-xs">
                <Globe className="w-5 h-5 text-zinc-600 dark:text-zinc-300" />
              </div>
              <div className="flex-1">
                <h4 className="font-medium text-sm text-zinc-900 dark:text-zinc-100">
                  Timezone
                </h4>
                <div className="mt-1 relative w-full">
                  <select
                    value={timezone}
                    onChange={(e) => setTimezone(e.target.value)}
                    className="block w-full appearance-none rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-3 py-1.5 text-sm text-zinc-900 dark:text-white focus:outline-none pr-8 cursor-pointer"
                  >
                    <option value="UTC">UTC (Universal Time)</option>
                    <option value="Asia/Kolkata">
                      IST (Indian Standard Time)
                    </option>
                    <option value="America/New_York">
                      EST (Eastern Standard Time)
                    </option>
                    <option value="Europe/London">
                      GMT (Greenwich Mean Time)
                    </option>
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-zinc-500">
                    <ChevronDown className="w-4 h-4" aria-hidden="true" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* ================= SAVE BUTTON ================= */}
        <div className="pt-4 flex justify-end">
          <div className="w-full sm:w-auto">
            <SubmitButton
              loading={isPending}
              disabled={isPending}
              onClick={handleSave}
              className="px-6"
            >
              {isPending ? "Saving..." : "Save Preferences"}
            </SubmitButton>
          </div>
        </div>
      </div>
    </section>
  );
};


