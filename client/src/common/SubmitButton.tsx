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
        outline-none
        relative overflow-hidden group
        w-full flex items-center justify-center gap-2
        py-2.5 px-4 rounded-xl cursor-pointer
        bg-indigo-600 text-white font-medium text-[15px] tracking-wide
        border border-transparent
        transition-all duration-200 ease-in-out
        hover:bg-indigo-700
        disabled:opacity-65 disabled:cursor-not-allowed
        ${className}
      `}
    >
      {/* Shine Animation */}
      <div className="absolute inset-0 -translate-x-[150%] group-hover:translate-x-[150%] transition-transform duration-1000 ease-in-out bg-linear-to-r from-transparent via-white/20 to-transparent skew-x-[-25deg]" />

      {/* Content */}
      <span className="relative z-10 flex items-center justify-center gap-2.5">
        {loading && <Loader />}
        <span className={loading ? "opacity-90" : "opacity-100 text-zinc-50 text-md font-bold"}>
          {children}
        </span>
      </span>
    </button>
  );
}