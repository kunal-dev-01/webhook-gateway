import { PreferencesSection } from "./PreferencesSection";
import { ProfileSidePanel } from "./ProfileSidePanel";
import { SecuritySection } from "./SecuritySection";

const Settings = () => {

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      
      {/* Header */}
      <div className="mb-8 md:mb-12">
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
          Account Settings
        </h1>
        <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">
          Manage your profile, preferences, and security.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        
        {/* LEFT COLUMN: Sidebar */}
        <div className="lg:col-span-4 space-y-6 lg:sticky lg:top-8">
          <ProfileSidePanel />
        </div>

        <div className="lg:col-span-8 space-y-8">
          <SecuritySection />
          <PreferencesSection />
        </div>

      </div>
    </div>
  )
}

export default Settings