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
        w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200 flex items-center justify-center gap-2 disabled:opacity-70 disabled:pointer-events-none cursor-pointer
        ${className}
      `}
    >
      {loading && <Loader />}
      {children}
    </button>
  );
}