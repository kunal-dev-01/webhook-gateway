// src/pages/auth/Login.tsx

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useLogin, useVerify2FA } from '@/hooks/useAuth';
import AuthInput from '@/common/AuthInput';
import SubmitButton from '@/common/SubmitButton';
import Toast from '@/common/Toast';
import { Lock, User } from '@/svg/Auth';
import { ChevronDown } from '@/svg/Common';
import type { LoginRequest, Verify2FARequest } from '@/types/auth.types';
import { loginSchema, twoFactorSchema, type LoginFormData, type TwoFactorFormData } from '@/schemas/auth.schema';
import { Link } from 'react-router-dom';



// ============================================================================
// Login Component
// ============================================================================

const Login = () => {
  // Toast state
  const [toastStatus, setToastStatus] = useState<'success' | 'error' | null>(null);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // 2FA state
  const [showOTPForm, setShowOTPForm] = useState(false);
  const [tempToken, setTempToken] = useState<string | null>(null);
  const [useRecoveryCode, setUseRecoveryCode] = useState(false);
  const [userEmail, setUserEmail] = useState<string>('');

  // Forms
  const loginForm = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const otpForm = useForm<TwoFactorFormData>({
    resolver: zodResolver(twoFactorSchema),
    defaultValues: {
      code: '',
    },
  });

  // Mutations
  const loginMutation = useLogin();
  const verify2FAMutation = useVerify2FA();

  // ============================================================================
  // Effect: Handle Login Response
  // ============================================================================
  useEffect(() => {
    if (loginMutation.isSuccess && loginMutation.data) {
      const data = loginMutation.data;

      // Case 1: 2FA Required
      if (data.require_2fa && data.temp_token) {
        setTempToken(data.temp_token);
        setShowOTPForm(true);
        setToastStatus('success');
        setToastMessage('Please enter your 2FA code to continue');
      } 
      // Case 2: Direct Login Success (No 2FA)
      else if (data.access_token && data.user) {
        setToastStatus('success');
        setToastMessage('Login successful! Redirecting...');
        setTimeout(() => {
          window.location.href = '/dashboard';
        }, 500);
      }
    }
  }, [loginMutation.isSuccess, loginMutation.data]);

  // ============================================================================
  // Effect: Handle 2FA Verification Success
  // ============================================================================
  useEffect(() => {
    if (verify2FAMutation.isSuccess && verify2FAMutation.data) {
      setToastStatus('success');
      setToastMessage('2FA verified! Redirecting...');
      setTimeout(() => {
        window.location.href = '/dashboard';
      }, 500);
    }
  }, [verify2FAMutation.isSuccess, verify2FAMutation.data]);

  // ============================================================================
  // Effect: Handle Login Errors
  // ============================================================================
  useEffect(() => {
    if (loginMutation.isError && loginMutation.error) {
      const errorMessage = 
        (loginMutation.error as any)?.response?.data?.error || 
        'Login failed. Please check your credentials.';
      setToastStatus('error');
      setToastMessage(errorMessage);
    }
  }, [loginMutation.isError, loginMutation.error]);

  // ============================================================================
  // Effect: Handle 2FA Verification Errors
  // ============================================================================
  useEffect(() => {
    if (verify2FAMutation.isError && verify2FAMutation.error) {
      const errorMessage = 
        (verify2FAMutation.error as any)?.response?.data?.error || 
        'Invalid code. Please try again.';
      setToastStatus('error');
      setToastMessage(errorMessage);
    }
  }, [verify2FAMutation.isError, verify2FAMutation.error]);

  // ============================================================================
  // Handlers
  // ============================================================================

  const onLoginSubmit = (data: LoginFormData) => {
    setUserEmail(data.email);
    loginMutation.login(data as LoginRequest);
  };

  const onOTPSubmit = (data: TwoFactorFormData) => {
    if (!tempToken) {
      setToastStatus('error');
      setToastMessage('Session expired. Please login again.');
      handleBackToLogin();
      return;
    }

    const request: Verify2FARequest = {
      temp_token: tempToken,
      code: data.code.replace(/\s/g, '').toUpperCase(), // Remove spaces and uppercase
    };

    verify2FAMutation.verify2FA(request);
  };

  const handleBackToLogin = () => {
    setShowOTPForm(false);
    setTempToken(null);
    setUseRecoveryCode(false);
    otpForm.reset();
    loginForm.reset();
    loginMutation.reset();
    verify2FAMutation.reset();
  };

  const handleToastClose = () => {
    setToastStatus(null);
    setToastMessage(null);
  };

  const toggleRecoveryCode = () => {
    setUseRecoveryCode(!useRecoveryCode);
    otpForm.setValue('code', '');
    otpForm.clearErrors('code');
  };

  const isLoading = loginMutation.isPending || verify2FAMutation.isPending;

  // ============================================================================
  // Render
  // ============================================================================

  return (
    <div className="w-full max-w-md mx-auto">
      {/* ========================================================================
          Header
      ======================================================================== */}
      <div className="text-center mb-8">
        <div className={`w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 transition-all duration-300 ${
          showOTPForm 
            ? 'bg-indigo-100 dark:bg-indigo-900/30' 
            : 'bg-linear-to-br from-indigo-500 to-purple-600'
        }`}>
          {showOTPForm ? (
            <Lock />
          ) : (
            <User className="w-8 h-8 text-white" />
          )}
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

      {/* ========================================================================
          Toast Notifications
      ======================================================================== */}
      <Toast status={toastStatus} message={toastMessage} onClose={handleToastClose} />

      {/* ========================================================================
          Login Form (Email + Password)
      ======================================================================== */}
      {!showOTPForm && (
        <form onSubmit={loginForm.handleSubmit(onLoginSubmit)} className="space-y-5">
          <AuthInput
            label="Email Address"
            type="email"
            name="email"
            placeholder="Enter your email"
            register={loginForm.register}
            error={loginForm.formState.errors.email?.message}
            disabled={isLoading}
          />

          <AuthInput
            label="Password"
            type="password"
            name="password"
            placeholder="Enter your password"
            register={loginForm.register}
            error={loginForm.formState.errors.password?.message}
            disabled={isLoading}
          />

          {/* Forgot Password Link */}
          <div className="flex justify-end">
            <Link
              to="/auth/forgot-password"
              aria-label='forgot password'
              className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors hover:underline"
            >
              Forgot password?
            </Link>
          </div>

          <SubmitButton loading={loginMutation.isPending} type="submit">
            {loginMutation.isPending ? 'Signing in...' : 'Sign In'}
          </SubmitButton>

          {/* Divider */}
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

          {/* Register Link */}
          <Link aria-label='create new account' to="/auth/register">
            <button
              type="button"
              className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200"
            >
              Create New Account
            </button>
          </Link>
        </form>
      )}

      {/* ========================================================================
          2FA Verification Form (TOTP or Recovery Code)
      ======================================================================== */}
      {showOTPForm && (
        <form onSubmit={otpForm.handleSubmit(onOTPSubmit)} className="space-y-5 animate-in fade-in slide-in-from-top-4 duration-300">
          {/* User Email Display */}
          {userEmail && (
            <div className="bg-zinc-50 dark:bg-zinc-800/50 rounded-xl p-3 border border-zinc-200 dark:border-zinc-700">
              <p className="text-xs text-zinc-500 dark:text-zinc-400">Signing in as</p>
              <p className="text-sm font-medium text-zinc-900 dark:text-white">{userEmail}</p>
            </div>
          )}

          {/* Code Type Info */}
          <div className="bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-900 rounded-xl p-3">
            <div className="flex items-start gap-2">
              <svg className="w-5 h-5 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
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

          {/* Code Input */}
          <AuthInput
            label={useRecoveryCode ? 'Recovery Code' : 'Verification Code'}
            type="text"
            name="code"
            placeholder={useRecoveryCode ? 'ABCD1234' : '123456'}
            register={otpForm.register}
            error={otpForm.formState.errors.code?.message}
            disabled={isLoading}
          />

          {/* Submit Button */}
          <SubmitButton loading={verify2FAMutation.isPending} type="submit">
            {verify2FAMutation.isPending ? 'Verifying...' : 'Verify & Login'}
          </SubmitButton>

          {/* Toggle Between TOTP and Recovery Code */}
          <div className="text-center">
            <button
              type="button"
              onClick={toggleRecoveryCode}
              className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors flex items-center gap-1 mx-auto group"
            >
              <svg className="w-4 h-4 group-hover:rotate-180 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {useRecoveryCode ? 'Use authenticator code instead' : 'Use recovery code instead'}
            </button>
          </div>

          {/* Back to Login */}
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
      )}

      {/* ========================================================================
          Footer
      ======================================================================== */}
      {!showOTPForm && (
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
      )}

      {/* Footer for 2FA */}
      {showOTPForm && (
        <div className="mt-6 space-y-3">
          {/* Warning for Recovery Code */}
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

          {/* Support Link */}
          <p className="text-center text-xs text-zinc-500 dark:text-zinc-500">
            Can't access your codes?{' '}
            <Link target="_blank" aria-label='get support' to="https://upnixo.com/support" className="text-indigo-600 dark:text-indigo-400 hover:underline">
              Get help
            </Link>
          </p>
        </div>
      )}
    </div>
  );
};

export default Login;