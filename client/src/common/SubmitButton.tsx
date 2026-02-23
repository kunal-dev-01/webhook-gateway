import React from 'react';
import type { ButtonHTMLAttributes } from 'react';
import { Loader } from './Loader';

type Props = {
  loading?: boolean;
  children: React.ReactNode;
} & ButtonHTMLAttributes<HTMLButtonElement>;

export default function SubmitButton({
  loading = false,
  children,
  className = '',
  ...rest
}: Props) {
  return (
    <button
      disabled={loading || rest.disabled}
      {...rest}
      className={`
        relative overflow-hidden group
        w-full flex items-center justify-center gap-2
        py-2.5 px-4 rounded-2xl cursor-pointer
        bg-indigo-600 text-white font-medium text-[15px] tracking-wide
        border border-transparent
        transition-colors duration-200 ease-in-out
        hover:bg-indigo-700
        focus:outline-none
        disabled:opacity-60 disabled:cursor-not-allowed
        ${className}
      `}
    >
      <div className="absolute inset-0 -translate-x-[150%] group-hover:translate-x-[150%] transition-transform duration-1000 ease-in-out bg-linear-to-r from-transparent via-white/15 to-transparent skew-x-[-25deg]" />

      {/* Content */}
      <span className="relative z-10 flex items-center gap-2">
        {loading && <Loader className="w-5 h-5 text-white/90" />}
        {children}
      </span>
    </button>
  );
}