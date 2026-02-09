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
        w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-70 disabled:pointer-events-none cursor-pointer
        ${className}
      `}
    >
      {/* Shine Effect Div */}
      <div className="absolute inset-0 -translate-x-full group-hover:translate-x-full transition-transform duration-700 ease-in-out bg-linear-to-r from-transparent via-white/25 to-transparent skew-x-[-20deg]" />

      {/* Content (Z-index added to keep text above shine) */}
      <span className="relative z-10 flex items-center gap-2">
        {loading && <Loader className='w-5 h-5 text-white' />}
        {children}
      </span>
    </button>
  );
}