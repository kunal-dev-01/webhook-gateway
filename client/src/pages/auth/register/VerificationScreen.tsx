import { Link } from "react-router-dom";
import { ChevronDown, Mail } from "@/svg/Common";

type Props = {
  registeredEmail: string;
  cooldown: number;
  isPending: boolean;
  onResend: () => void;
};

export default function VerificationScreen({ registeredEmail, cooldown, isPending, onResend }: Props) {
  return (
    <div className="w-full max-w-md mx-auto animate-in fade-in zoom-in-95 duration-300">
      <div className="text-center mb-8">
        <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
          <Mail className="w-10 h-10 text-indigo-500" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">Check Your Email</h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-4">
          We've sent a verification link to <br />
          <span className="font-semibold text-zinc-900 dark:text-white">{registeredEmail}</span>
        </p>
      </div>
      
      <button
        onClick={onResend}
        disabled={cooldown > 0 || isPending}
        className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed mb-4"
      >
        {isPending ? "Sending..." : cooldown > 0 ? `Resend in ${cooldown}s` : "Resend Verification Email"}
      </button>

      <div className="text-center">
        <Link to="/auth/login" className="text-sm text-indigo-600 dark:text-indigo-400 flex items-center justify-center gap-1 hover:underline">
          <ChevronDown className="w-4 h-4 rotate-90" /> Back to login
        </Link>
      </div>
    </div>
  );
}