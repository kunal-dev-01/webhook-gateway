import { Link } from 'react-router-dom';

type Props = {
  showOTPForm: boolean;
  useRecoveryCode: boolean;
};

export default function LoginFooter({ showOTPForm, useRecoveryCode }: Props) {
  if (!showOTPForm) {
    return (
      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500 mt-6">
        By continuing, you agree to our{' '}
        <Link aria-label='terms of service' to="https://upnixo.com/legal/terms-of-service" target='_blank' className="text-indigo-600 dark:text-indigo-400 hover:underline">
          Terms of Service
        </Link>{' '}  
        and{' '}
        <Link aria-label='privacy policy' to="https://upnixo.com/legal/privacy-policy" target='_blank' className="text-indigo-600 dark:text-indigo-400 hover:underline">
          Privacy Policy
        </Link>
      </p>
    );
  }

  return (
    <div className="mt-6 space-y-3">
      {useRecoveryCode && (
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-900 rounded-xl p-3">
          <div className="flex items-start gap-2">
            <svg className="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div className="text-xs">
              <p className="font-medium text-amber-900 dark:text-amber-200">Important</p>
              <p className="text-amber-700 dark:text-amber-300 mt-0.5">
                This recovery code will be permanently deleted after use. Make sure you have other codes saved.
              </p>
            </div>
          </div>
        </div>
      )}
      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500">
        Can't access your codes?{' '}
        <Link target="_blank" aria-label='get support' to="https://upnixo.com/support" className="text-indigo-600 dark:text-indigo-400 hover:underline">
          Get help
        </Link>
      </p>
    </div>
  );
}