import { Loader } from "@/common/Loader";

export default function VerifyingScreen() {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      <div className="text-center">
        {/* Loading Icon */}
        <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
          <Loader className="w-10 h-10 text-indigo-600 dark:text-indigo-400" />
        </div>

        {/* Title */}
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
          Verifying Your Email
        </h1>

        {/* Description */}
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
          Please wait while we verify your email address...
        </p>

        {/* Loading Animation */}
        <div className="flex justify-center gap-2">
          <div
            className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
            style={{ animationDelay: "0ms" }}
          />
          <div
            className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
            style={{ animationDelay: "150ms" }}
          />
          <div
            className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
            style={{ animationDelay: "300ms" }}
          />
        </div>
      </div>
    </div>
  );
}