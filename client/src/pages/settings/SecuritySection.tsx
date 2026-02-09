import { ChangePasswordForm } from './ChangePasswordForm';
import { TwoFactorAuth } from './TwoFactorAuth';

export const SecuritySection = () => {
  return (
    <div className="space-y-8">
      
      {/* 2FA Section - High Priority */}
      <section className="bg-white dark:bg-zinc-900/80 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6">
         <TwoFactorAuth />
      </section>

      {/* Password Change Section */}
      <section className="bg-white dark:bg-zinc-900/80 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6">
        <ChangePasswordForm />
      </section>

    </div>
  );
};