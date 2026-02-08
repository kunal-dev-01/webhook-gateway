import { Outlet, Link } from "react-router-dom";
import { memo } from "react";
import { ChevronDown } from "@/svg/Common";

const AuthLayout = memo(() => {
  return (
    <div className="min-h-screen w-full flex items-center justify-center p-4 relative overflow-hidden transition-colors duration-500">
      {/* ─── Premium Gradient Backdrop ─── */}
      <div className="lg:block hidden absolute inset-0 overflow-hidden pointer-events-none">
        {/* Top Right Glow */}
        <div className="absolute -top-[10%] -right-[5%] w-[50%] h-[50%] rounded-full bg-indigo-500/10 dark:bg-indigo-500/15 blur-[120px]" />
        {/* Bottom Left Glow */}
        <div className="absolute -bottom-[10%] -left-[5%] w-[50%] h-[50%] rounded-full bg-indigo-600/10 dark:bg-indigo-600/15 blur-[120px]" />
      </div>

      {/* ─── Top Navigation (Back Button) ─── */}
      <Link
        to="/dashboard"
        className="absolute top-6 left-4 flex items-center gap-2 text-slate-500 dark:text-slate-400 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors duration-200 group cursor-pointer"
      >
        <ChevronDown className="w-4 h-4 transform rotate-90 group-hover:-translate-x-1 transition-transform duration-200" />
        <span className="text-sm font-semibold tracking-wide">Home</span>
      </Link>

      {/* ─── Main Content Card ─── */}
      <main className="w-full max-w-md z-10">
        <Outlet />
      </main>
    </div>
  );
});

export default AuthLayout;