import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner'; 
import { useLogin, useVerify2FA } from '@/hooks/useAuth';
import type { LoginRequest, Verify2FARequest } from '@/types/auth.types';
import { ApiError } from '@/types/api.types';
import { 
  loginSchema, 
  twoFactorSchema, 
  type LoginFormData, 
  type TwoFactorFormData 
} from '@/schemas/auth.schema';

// Import split components
import LoginHeader from './LoginHeader';
import LoginForm from './LoginForm';
import TwoFactorForm from './TwoFactorForm';
import LoginFooter from './LoginFooter';

const Login = () => {
  const [showOTPForm, setShowOTPForm] = useState(false);
  const [tempToken, setTempToken] = useState<string | null>(null);
  const [useRecoveryCode, setUseRecoveryCode] = useState(false);
  const [userEmail, setUserEmail] = useState<string>('');

  const loginForm = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const otpForm = useForm<TwoFactorFormData>({
    resolver: zodResolver(twoFactorSchema),
    defaultValues: { code: '' },
  });

  const loginMutation = useLogin();
  const verify2FAMutation = useVerify2FA();

  // Effects
  useEffect(() => {
    if (loginMutation.isSuccess && loginMutation.data) {
      const data = loginMutation.data;
      if (data.two_factor_required && data.temp_token) {
        setTempToken(data.temp_token);
        setShowOTPForm(true);
        toast.success('Please enter your 2FA code to continue');
      } else if (data.access_token && data.user) {
        toast.success('Login successful! Redirecting...');
        setTimeout(() => { window.location.href = '/dashboard'; }, 500);
      }
    }
  }, [loginMutation.isSuccess, loginMutation.data]);

  useEffect(() => {
    if (verify2FAMutation.isSuccess && verify2FAMutation.data) {
      toast.success('2FA verified! Redirecting...');
      setTimeout(() => { window.location.href = '/dashboard'; }, 500);
    }
  }, [verify2FAMutation.isSuccess, verify2FAMutation.data]);

  useEffect(() => {
    if (loginMutation.isError && loginMutation.error) {
      const error = loginMutation.error as ApiError;
      if (error.isValidationError()) {
        const emailError = error.getFieldError('email');
        const passwordError = error.getFieldError('password');
        if (emailError) loginForm.setError('email', { message: emailError });
        if (passwordError) loginForm.setError('password', { message: passwordError });
        toast.error(error.message || 'Validation failed');
      } else {
        toast.error(error.message || 'Login failed. Please check your credentials.');
      }
    }
  }, [loginMutation.isError, loginMutation.error, loginForm]);

  useEffect(() => {
    if (verify2FAMutation.isError && verify2FAMutation.error) {
      const error = verify2FAMutation.error as ApiError;
      if (error.isValidationError()) {
        const codeError = error.getFieldError('code');
        if (codeError) otpForm.setError('code', { message: codeError });
        toast.error(error.message || 'Invalid code format');
      } else {
        const message = error.isStatus(401)
          ? 'Invalid or expired code. Please try again.'
          : error.message || 'Verification failed. Please try again.';
        toast.error(message);
      }
    }
  }, [verify2FAMutation.isError, verify2FAMutation.error, otpForm]);

  // Handlers
  const onLoginSubmit = (data: LoginFormData) => {
    setUserEmail(data.email);
    loginMutation.login(data as LoginRequest);
  };

  const onOTPSubmit = (data: TwoFactorFormData) => {
    if (!tempToken) {
      toast.error('Session expired. Please login again.');
      handleBackToLogin();
      return;
    }
    const request: Verify2FARequest = {
      temp_token: tempToken,
      code: data.code.replace(/\s/g, '').toUpperCase(),
      is_recovery: useRecoveryCode,
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

  const toggleRecoveryCode = () => {
    setUseRecoveryCode(!useRecoveryCode);
    otpForm.setValue('code', '');
    otpForm.clearErrors('code');
  };

  const isLoading = loginMutation.isPending || verify2FAMutation.isPending;

  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      <LoginHeader showOTPForm={showOTPForm} useRecoveryCode={useRecoveryCode} />

      {!showOTPForm ? (
        <LoginForm 
          form={loginForm} 
          onSubmit={onLoginSubmit} 
          isLoading={isLoading} 
        />
      ) : (
        <TwoFactorForm 
          form={otpForm}
          onSubmit={onOTPSubmit}
          isLoading={isLoading}
          userEmail={userEmail}
          useRecoveryCode={useRecoveryCode}
          toggleRecoveryCode={toggleRecoveryCode}
          handleBackToLogin={handleBackToLogin}
        />
      )}

      <LoginFooter showOTPForm={showOTPForm} useRecoveryCode={useRecoveryCode} />
    </div>
  );
};

export default Login;