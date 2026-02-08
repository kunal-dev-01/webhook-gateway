import { EyeClose, EyeOpen } from "@/svg/Common";
import { memo, useState } from "react";

interface AuthInputProps {
  label: string;
  type: string;
  placeholder: string;
  register: any; 
  error?: string;
  name: string;
}

const AuthInput = memo(({ label, type, placeholder, register, error, name }: AuthInputProps) => {
  const [showPassword, setShowPassword] = useState(false);
  const isPassword = type === "password";

  return (
    <div className="flex flex-col gap-1.5 w-full">
      <label htmlFor={name} className="text-sm font-medium text-zinc-700 dark:text-zinc-200 ml-1">
        {label}
      </label>
      <div className="relative group">
        <input
          type={isPassword ? (showPassword ? "text" : "password") : type}
          placeholder={placeholder}
          {...register(name)}
          className={`w-full px-4 py-2.5 bg-transparent border rounded-xl outline-none transition-all duration-200
            ${error 
              ? "border-red-500 focus:ring-2 focus:ring-red-200 dark:focus:ring-red-900/30" 
              : "border-zinc-200 dark:border-zinc-800 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/10"
            } 
            dark:text-white text-zinc-900 placeholder:text-zinc-400 dark:placeholder:text-zinc-600`}
        />
        
        {isPassword && (
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-zinc-400 hover:text-indigo-500 transition-colors cursor-pointer"
          >
            {showPassword ? <EyeOpen /> : <EyeClose />}
          </button>
        )}
      </div>
      {error && <span className="text-xs text-red-500 ml-1 font-medium">{error}</span>}
    </div>
  );
});

export default AuthInput;