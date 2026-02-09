import { Credits, Dashboard } from "@/svg/Auth";
import { Settings } from "@/svg/Common";
import { memo, useMemo, type FC } from "react";
import { Outlet, useNavigate, useLocation } from "react-router-dom";

interface NavItemConfig {
  label: string;
  path: string;
  icon: FC<React.SVGProps<SVGSVGElement>>;
}

interface NavItemProps {
  item: NavItemConfig;
  isActive: boolean;
  onClick: (path: string) => void;
}


// --- NavItem Component ---
const NavItem = memo(({ item, isActive, onClick }: NavItemProps) => {
  const Icon = item.icon;

  return (
    <button
      onClick={() => onClick(item.path)}
      className={`
        group relative flex items-center justify-center px-4 py-2 rounded-full cursor-pointer outline-none
        ${isActive ? "text-indigo-600 dark:text-indigo-400" : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200"}
      `}
    >
      {/* Background Indicator Layer - Only this scales */}
      <div 
        className={`
          absolute inset-0 rounded-full -z-10 transition-all duration-300 ease-[cubic-bezier(0.23,1,0.32,1)]
          ${isActive 
            ? "bg-indigo-500/10 dark:bg-indigo-400/10 scale-100 opacity-100" 
            : "scale-75 opacity-0 group-hover:opacity-100 group-hover:scale-100 bg-zinc-100 dark:bg-zinc-800/80"
          }
        `} 
      />

      {/* Icon - No Scaling, Only Color Transition */}
      <Icon className="w-5.5 h-5.5 z-10 transition-colors duration-200" />

      {/* Tooltip */}
      <span className="absolute -top-10 scale-0 group-hover:scale-100 transition-all duration-200 bg-zinc-900 dark:bg-zinc-100 text-zinc-100 dark:text-zinc-900 text-[10px] font-medium px-2.5 py-1 rounded-full pointer-events-none shadow-sm">
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
    { label: "Dashboard", path: "/dashboard", icon: Dashboard },
    { label: "Credits", path: "/credits", icon: Credits },
    { label: "Settings", path: "/settings", icon: Settings },
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