import { memo, useMemo, type FC } from "react";
import { Outlet, useNavigate, useLocation } from "react-router-dom";

// --- Types ---
interface IconProps {
  className?: string;
}

interface NavItemConfig {
  label: string;
  path: string;
  icon: FC<IconProps>;
}

interface NavItemProps {
  item: NavItemConfig;
  isActive: boolean;
  onClick: (path: string) => void;
}

// --- Icons (Same logic, simple & clean) ---
const Icons = {
  Dashboard: ({ className }: IconProps) => (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="1.5" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M3 3h7v7H3z"/><path d="M14 3h7v7h-7z"/><path d="M14 14h7v7h-7z"/><path d="M3 14h7v7H3z"/>
    </svg>
  ),
  Settings: ({ className }: IconProps) => (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="1.5" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.09a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15-.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.09a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.1a2 2 0 0 1-1-1.72v-.51a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
      <circle cx="12" cy="12" r="3"/>
    </svg>
  ),
  Credits: ({ className }: IconProps) => (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="1.5" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="M17.3 15.3l.6.7 1.3-2"/>
    </svg>
  ),
  Payments: ({ className }: IconProps) => (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="1.5" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/>
    </svg>
  ),
  History: ({ className }: IconProps) => (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="1.5" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/>
    </svg>
  ),
};

// --- NavItem Component ---
const NavItem = memo(({ item, isActive, onClick }: NavItemProps) => {
  const Icon = item.icon;
  return (
    <button
      onClick={() => onClick(item.path)}
      className={`
        group relative flex items-center justify-center px-4 py-2 rounded-full transition-all duration-200 cursor-pointer
        ${isActive 
          ? "text-indigo-600 bg-indigo-500/10 dark:bg-indigo-400/10" 
          : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100 hover:bg-zinc-100 dark:hover:bg-zinc-800/80"}
      `}
    >
      <Icon 
        className={`
          w-6 h-6 transition-transform duration-200 
          ${isActive ? "scale-110" : ""}
        `} 
      />
      {/* Tooltip purely for accessibility/UI feedback */}
      <span className="absolute -top-12 scale-0 group-hover:scale-100 transition-all bg-zinc-900 dark:bg-zinc-100 text-zinc-100 dark:text-zinc-900 text-[10px] font-medium px-3 py-1 rounded-full">
        {item.label}
      </span>
    </button>
  );
});

// --- Main Layout ---
const DashboardLayout: FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const navItems = useMemo<NavItemConfig[]>(() => [
    { label: "Dashboard", path: "/dashboard", icon: Icons.Dashboard },
    { label: "History", path: "/history", icon: Icons.History },
    { label: "Payments", path: "/payments", icon: Icons.Payments },
    { label: "Credits", path: "/credits", icon: Icons.Credits },
    { label: "Settings", path: "/settings", icon: Icons.Settings },
  ], []);

  const isPathActive = (path: string): boolean => {
    return location.pathname.startsWith(path);
  };

  return (
    <div className="min-h-screen w-full flex flex-col items-center max-w-7xl mx-auto relative">
      
      {/* Content */}
      <main className="w-full flex-1 p-4 pb-28">
        <Outlet />
      </main>

      {/* Floating Thin Nav Bar */}
      <div className="fixed bottom-6 z-50 px-4 w-full flex justify-center pointer-events-none">
        <nav className="
            pointer-events-auto
            flex items-center justify-between sm:justify-center 
            gap-2 sm:gap-8 
            px-4 py-2.5 
            w-full sm:w-max
            rounded-full
            backdrop-blur-md 
            bg-white/70 border border-zinc-200/50
            dark:bg-zinc-900/60 dark:border-zinc-800/50 
          ">
          {navItems.map((item) => (
            <NavItem 
              key={item.label} 
              item={item} 
              isActive={isPathActive(item.path)}
              onClick={(path) => navigate(path)}
            />
          ))}
        </nav>
      </div>
      
    </div>
  );
};

export default memo(DashboardLayout);