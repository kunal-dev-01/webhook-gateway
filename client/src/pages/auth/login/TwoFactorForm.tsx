import { type UseFormReturn } from 'react-hook-form';
import AuthInput from '@/common/AuthInput';
import SubmitButton from '@/common/SubmitButton';
import { ChevronDown, Refresh, Error } from '@/svg/Common';
import { type TwoFactorFormData } from '@/schemas/auth.schema';

type Props = {
  form: UseFormReturn<TwoFactorFormData>;
  onSubmit: (data: TwoFactorFormData) => void;
  isLoading: boolean;
  userEmail: string;
  useRecoveryCode: boolean;
  toggleRecoveryCode: () => void;
  handleBackToLogin: () => void;
};

export default function TwoFactorForm({
  form, onSubmit, isLoading, userEmail, useRecoveryCode, toggleRecoveryCode, handleBackToLogin
}: Props) {
  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5 anim-fade-in">
      {userEmail && (
        <div className="bg-zinc-50 dark:bg-zinc-800/50 rounded-xl p-3 border border-zinc-200 dark:border-zinc-700">
          <p className="text-xs text-zinc-500 dark:text-zinc-400">Signing in as</p>
          <p className="text-sm font-medium text-zinc-900 dark:text-white">{userEmail}</p>
        </div>
      )}

      <div className="bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-900 rounded-xl p-3">
        <div className="flex items-start gap-2">
          <Error className="w-5 h-5 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" />
          <div className="text-sm">
            <p className="font-medium text-indigo-900 dark:text-indigo-200">
              {useRecoveryCode ? 'Using Recovery Code' : 'Using Authenticator App'}
            </p>
            <p className="text-xs text-indigo-700 dark:text-indigo-300 mt-0.5">
              {useRecoveryCode 
                ? 'Enter 8-character code (e.g., ABCD1234). This code will be deleted after use.' 
                : 'Enter the 6-digit code from your authenticator app.'}
            </p>
          </div>
        </div>
      </div>

      <AuthInput
        label={useRecoveryCode ? 'Recovery Code' : 'Verification Code'}
        type="text"
        name="code"
        placeholder={useRecoveryCode ? 'ABCD1234' : '123456'}
        register={form.register}
        error={form.formState.errors.code?.message}
        disabled={isLoading}
      />

      <SubmitButton loading={isLoading} type="submit">
        {isLoading ? 'Verifying...' : 'Verify & Login'}
      </SubmitButton>

      <div className="text-center">
        <button
          type="button"
          onClick={toggleRecoveryCode}
          className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors flex items-center gap-1 mx-auto group"
        >
          <Refresh className="w-4.5 h-4.5 group-hover:rotate-180 transition-transform duration-300" />
          {useRecoveryCode ? 'Use authenticator code instead' : 'Use recovery code instead'}
        </button>
      </div>

      <div className="text-center pt-4 border-t border-zinc-200 dark:border-zinc-800">
        <button
          type="button"
          onClick={handleBackToLogin}
          className="text-sm text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200 transition-colors flex items-center justify-center mx-auto gap-2"
        >
          <ChevronDown className="w-4 h-4 rotate-90" /> Back to login
        </button>
      </div>
    </form>
  );
}