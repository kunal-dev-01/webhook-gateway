import React from "react";

interface RadioButtonProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  wrapperClassName?: string;
}

const RadioButton = ({
  label,
  id,
  className = "",
  wrapperClassName = "",
  disabled,
  ...props
}: RadioButtonProps) => {
  return (
    <label
      htmlFor={id}
      className={`group flex items-center gap-3 flex-1 relative p-3 rounded-xl border border-transparent hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-all duration-200 ${
        disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"
      } ${wrapperClassName}`}
    >
      {/* 1. Hidden Native Input (Logic yahi sambhalega) */}
      <input
        type="radio"
        id={id}
        disabled={disabled}
        className="peer sr-only" // sr-only means visually hidden but accessible
        {...props}
      />

      {/* 2. Custom Radio Circle (Design ye dikhayega) */}
      <div
        className={`
          relative flex items-center justify-center w-5 h-5 rounded-full border-2 
          border-zinc-300 dark:border-zinc-600 bg-white dark:bg-zinc-900
          transition-colors duration-200
          
          /* Peer Checked State (Jab input check ho) */
          peer-checked:border-indigo-600 peer-checked:dark:border-indigo-500
          
          /* Group Hover State (Thoda highlight) */
          group-hover:border-indigo-400
          ${className}
        `}
      >
        {/* The Inner Dot (Animated) */}
        <div className="w-2.5 h-2.5 rounded-full bg-indigo-600 dark:bg-indigo-500 scale-0 peer-checked:scale-100 transition-transform duration-200 ease-out" />
      </div>

      {/* 3. Label Text */}
      <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300 select-none group-hover:text-zinc-900 dark:group-hover:text-white transition-colors">
        {label}
      </span>
      
      {/* Optional: Active Border around the whole box when checked */}
      <div className="absolute inset-0 rounded-xl border-2 border-indigo-600 opacity-0 peer-checked:opacity-10 pointer-events-none transition-opacity" />
    </label>
  );
};

export default RadioButton;