// src/components/common/PasswordStrength.tsx

import { memo, useMemo } from 'react';

interface PasswordStrengthProps {
  password: string;
}

const PasswordStrength = memo(({ password }: PasswordStrengthProps) => {
  const strength = useMemo(() => {
    if (!password) return { level: 0, text: '', color: '' };

    let score = 0;
    
    // Length check
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    
    // Character variety checks
    if (/[a-z]/.test(password)) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^A-Za-z0-9]/.test(password)) score++;

    // Determine strength level
    if (score <= 2) {
      return { level: 1, text: 'Weak', color: 'bg-red-500' };
    } else if (score <= 4) {
      return { level: 2, text: 'Fair', color: 'bg-orange-500' };
    } else if (score <= 5) {
      return { level: 3, text: 'Good', color: 'bg-yellow-500' };
    } else {
      return { level: 4, text: 'Strong', color: 'bg-green-500' };
    }
  }, [password]);

  if (!password) return null;

  return (
    <div className="space-y-2">
      {/* Strength Bars */}
      <div className="flex gap-1.5">
        {[1, 2, 3, 4].map((level) => (
          <div
            key={level}
            className={`h-1.5 flex-1 rounded-full transition-all duration-300 ${
              level <= strength.level
                ? strength.color
                : 'bg-zinc-200 dark:bg-zinc-800'
            }`}
          />
        ))}
      </div>

      {/* Strength Text */}
      <div className="flex items-center justify-between text-xs">
        <span className="text-zinc-600 dark:text-zinc-400">
          Password strength: <span className={`font-semibold ${
            strength.level === 1 ? 'text-red-500' :
            strength.level === 2 ? 'text-orange-500' :
            strength.level === 3 ? 'text-yellow-600 dark:text-yellow-500' :
            'text-green-600 dark:text-green-500'
          }`}>{strength.text}</span>
        </span>
      </div>

      {/* Requirements Checklist */}
      <div className="space-y-1 pt-1">
        <RequirementItem met={password.length >= 8} text="At least 8 characters" />
        <RequirementItem met={/[A-Z]/.test(password)} text="One uppercase letter" />
        <RequirementItem met={/[a-z]/.test(password)} text="One lowercase letter" />
        <RequirementItem met={/[0-9]/.test(password)} text="One number" />
        <RequirementItem met={/[^A-Za-z0-9]/.test(password)} text="One special character" />
      </div>
    </div>
  );
});

const RequirementItem = ({ met, text }: { met: boolean; text: string }) => (
  <div className="flex items-center gap-2 text-xs">
    <div className={`w-4 h-4 rounded-full flex items-center justify-center transition-colors ${
      met ? 'bg-green-100 dark:bg-green-900/30' : 'bg-zinc-100 dark:bg-zinc-800'
    }`}>
      {met ? (
        <svg className="w-3 h-3 text-green-600 dark:text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <div className="w-1.5 h-1.5 rounded-full bg-zinc-400 dark:bg-zinc-600" />
      )}
    </div>
    <span className={met ? 'text-green-600 dark:text-green-500' : 'text-zinc-500 dark:text-zinc-500'}>
      {text}
    </span>
  </div>
);

export default PasswordStrength;