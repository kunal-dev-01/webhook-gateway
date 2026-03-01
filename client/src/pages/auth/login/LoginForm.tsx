import { type UseFormReturn } from 'react-hook-form';
import { Link } from 'react-router-dom';
import AuthInput from '@/common/AuthInput';
import SubmitButton from '@/common/SubmitButton';
import { type LoginFormData } from '@/schemas/auth.schema';

type Props = {
  form: UseFormReturn<LoginFormData>;
  onSubmit: (data: LoginFormData) => void;
  isLoading: boolean;
};

export default function LoginForm({ form, onSubmit, isLoading }: Props) {
  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
      <AuthInput
        label="Email Address"
        type="email"
        name="email"
        placeholder="Enter your email"
        register={form.register}
        error={form.formState.errors.email?.message}
        disabled={isLoading}
      />

      <AuthInput
        label="Password"
        type="password"
        name="password"
        placeholder="Enter your password"
        register={form.register}
        error={form.formState.errors.password?.message}
        disabled={isLoading}
      />

      <div className="flex justify-end">
        <Link
          to="/auth/forgot-password"
          aria-label='forgot password'
          className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors hover:underline"
        >
          Forgot password?
        </Link>
      </div>

      <SubmitButton loading={isLoading} type="submit">
        {isLoading ? 'Signing in...' : 'Sign In'}
      </SubmitButton>

      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-zinc-200 dark:border-zinc-800"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-4 bg-white dark:bg-zinc-900 text-zinc-500 dark:text-zinc-400 rounded-full">
            Don't have an account?
          </span>
        </div>
      </div>

      <Link aria-label='create new account' to="/auth/register">
        <button
          type="button"
          className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200"
        >
          Create New Account
        </button>
      </Link>
    </form>
  );
}