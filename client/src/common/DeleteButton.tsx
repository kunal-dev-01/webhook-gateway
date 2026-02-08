import React from 'react';
import type { ButtonHTMLAttributes } from 'react';
import { Loader } from './Loader';

type Props = {
  loading?: boolean;
  children: React.ReactNode;
} & ButtonHTMLAttributes<HTMLButtonElement>;

export default function DeleteButton({
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
        group relative overflow-hidden
        w-full bg-red-500 hover:bg-red-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-70 disabled:pointer-events-none cursor-pointer
        ${className}
      `}
    >
      {/* Shine Effect Element */}
      <div className="absolute top-0 -left-full w-full h-full bg-linear-to-r from-transparent via-white/25 to-transparent skew-x-[-20deg] group-hover:left-full transition-[left] duration-1000 ease-in-out" />

      {/* Content Wrapper */}
      <span className="relative z-10 flex items-center gap-2">
        {loading && <Loader className='w-5 h-5 text-white' />}
        {children}
      </span>
    </button>
  );
}