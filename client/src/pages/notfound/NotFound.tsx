import { useMemo } from "react";

const NotFound = () => {
  const referenceId = useMemo(
    () => `ERR_404_VOID_${Math.random().toString(36).slice(2, 11).toUpperCase()}`,
    []
  );

  return (
    <div className="min-h-screen flex items-center justify-center px-6">
      <div className="w-full max-w-2xl text-center space-y-10">
        
        {/* 404 Text */}
        <h1 className="text-[12rem] font-black leading-none tracking-tighter text-transparent bg-clip-text bg-linear-to-b from-zinc-200 to-zinc-800 dark:from-zinc-100 dark:opacity-20 opacity-50 select-none">
          404
        </h1>

        {/* Message */}
        <div className="space-y-4">
          <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-zinc-700 dark:text-zinc-200">
            Lost in the digital void?
          </h2>
          <p className="text-lg text-zinc-400 max-w-md mx-auto leading-relaxed">
            The page you're looking for has moved to another dimension or simply never existed.
          </p>
        </div>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row justify-center gap-4 pt-8">
          <button
            onClick={() => (window.location.href = "/dashboard")}
            className="px-10 py-2.5 rounded-xl font-bold bg-zinc-100 text-zinc-950 hover:bg-zinc-200 transition-all duration-200 ease-in-out cursor-pointer"
          >
            Back to Home
          </button>

          <button
            onClick={() => window.history.back()}
            className="px-10 py-2.5 rounded-xl border border-zinc-300 dark:border-zinc-800 dark:text-zinc-400 text-zinc-700 font-semibold hover:bg-zinc-700 hover:text-zinc-100 transition-all duration-300 ease-in-out cursor-pointer"
          >
            Previous Page
          </button>
        </div>

        {/* Footer */}
        <p className="pt-16 text-xs uppercase font-medium text-zinc-600">
          Reference ID:{" "}
          <span className="font-mono text-zinc-500">{referenceId}</span>
        </p>
      </div>
    </div>
  );
};

export default NotFound;