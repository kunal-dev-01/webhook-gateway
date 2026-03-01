import { Lock, User } from '@/svg/Auth';

type Props = {
  showOTPForm: boolean;
  useRecoveryCode: boolean;
};

export default function LoginHeader({ showOTPForm, useRecoveryCode }: Props) {
  return (
    <div className="text-center mb-8">
      <div className={`w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 transition-all duration-300 ${
        showOTPForm 
          ? 'bg-indigo-100 dark:bg-indigo-900/30' 
          : 'dark:bg-zinc-900 bg-zinc-100 border border-zinc-200 dark:border-zinc-800 '
      }`}>
        {showOTPForm ? <Lock className="w-10 h-10 text-indigo-600 dark:text-indigo-400" /> : <User className="w-12 h-12 text-zinc-900 dark:text-zinc-100" />}
      </div>
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
        {showOTPForm ? 'Two-Factor Authentication' : 'Welcome Back'}
      </h1>
      <p className="text-sm text-zinc-600 dark:text-zinc-400">
        {showOTPForm 
          ? (useRecoveryCode 
              ? 'Enter your 8-character recovery code' 
              : 'Enter the 6-digit code from your authenticator app')
          : 'Sign in to your account to continue'}
      </p>
    </div>
  );
}