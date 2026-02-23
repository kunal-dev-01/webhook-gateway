import { ChevronDown } from "@/svg/Common";
import { useState, useRef, useEffect } from "react";


interface Option {
  label: string;
  value: string;
}

interface SelectInputProps {
  label: string;
  options: Option[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  className?: string;
}

const SelectInput = ({
  label,
  options,
  value,
  onChange,
  disabled,
  className,
}: SelectInputProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close dropdown if clicked outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const selectedLabel = options.find((opt) => opt.value === value)?.label;

  return (
    <div className="relative flex-1" ref={containerRef}>
      <label htmlFor="select" className="block text-sm font-medium text-zinc-700 dark:text-zinc-200 ml-1 mb-1.5">
        {label}
      </label>

      {/* Trigger Button */}
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`w-full flex items-center justify-between p-2 border rounded-xl transition-all duration-200 outline-none ${
          isOpen
            ? "dark:border-indigo-500 border-indigo-400"
            : "border-zinc-200 dark:border-zinc-800 hover:border-indigo-400 dark:hover:border-indigo-500"
        } ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"} ${className}`}
      >
        <span className="text-sm font-medium text-zinc-900 dark:text-white">
          {selectedLabel || "Select Option"}
        </span>
        <ChevronDown
          className={`w-5 h-5 text-zinc-500 transition-transform duration-200 ${
            isOpen ? "rotate-180" : ""
          }`}
        />
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-2 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-xl overflow-hidden  duration-100">
          <ul className="max-h-60 overflow-auto py-1">
            {options.map((option) => (
              <li key={option.value}>
                <button
                  type="button"
                  onClick={() => {
                    onChange(option.value);
                    setIsOpen(false);
                  }}
                  className={`w-full text-left px-4 py-2.5 text-sm font-medium transition-colors ${
                    value === option.value
                      ? "bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400"
                      : "text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800"
                  }`}
                >
                  {option.label}
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default SelectInput;