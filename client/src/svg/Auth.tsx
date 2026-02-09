import { memo } from "react";

// ========================== User ===========================
export const User = memo(
  ({ className = "w-7 h-7 text-white" }: { className?: string }) => (
    <svg
      width="100%"
      height="100%"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <path
        d="M12.0672 2C9.6678 2 7.72266 3.94514 7.72266 6.34459C7.72266 8.74404 9.6678 10.6892 12.0672 10.6892C14.4667 10.6892 16.4118 8.74404 16.4118 6.34459C16.4118 3.94514 14.4667 2 12.0672 2Z"
        fill="currentColor"
      />
      <path
        d="M20.25 19.453C20.2421 19.8615 19.9087 20.1895 19.5001 20.1895H4.50013C4.09157 20.1895 3.75818 19.8624 3.75027 19.454L3.75023 19.452L3.75019 19.4493L3.7501 19.4423L3.75 19.4211C3.74999 19.404 3.75015 19.3807 3.75072 19.3516C3.75187 19.2937 3.75468 19.2127 3.76117 19.1119C3.77413 18.9107 3.80189 18.6292 3.86099 18.2937C3.97867 17.6258 4.22374 16.7262 4.73808 15.8194C5.79641 13.9537 7.92408 12.1895 12.0001 12.1895C16.0762 12.1895 18.2038 13.9537 19.2622 15.8194C19.7765 16.7262 20.0216 17.6258 20.1393 18.2937C20.1984 18.6292 20.2261 18.9107 20.2391 19.1119C20.2456 19.2127 20.2484 19.2937 20.2495 19.3516C20.2501 19.3807 20.2503 19.404 20.2503 19.4211L20.2502 19.4423L20.2501 19.4493L20.25 19.452L20.25 19.453Z"
        fill="currentColor"
      />
    </svg>
  ),
);

// ========================== Key ===========================
export const Key = memo(
  ({ className = "w-8 h-8 text-white" }: { className?: string }) => (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
      />
    </svg>
  ),
);

// ================= Lock ================== //
export const Lock = memo(
  ({
    className = "w-8 h-8 text-indigo-600 dark:text-indigo-400",
  }: {
    className?: string;
  }) => (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
      />
    </svg>
  ),
);

// ================= Dashboard ================== //
export const Dashboard = memo(({ className }: { className?: string }) => (
  <svg
    className={className}
    width="100%"
    height="100%"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      fillRule="evenodd"
      clipRule="evenodd"
      d="M3 2C2.44772 2 2 2.44772 2 3V6C2 6.55228 2.44772 7 3 7H6C6.55228 7 7 6.55228 7 6V3C7 2.44772 6.55228 2 6 2H3ZM3 0H6C7.65685 0 9 1.34315 9 3V6C9 7.65685 7.65685 9 6 9H3C1.34315 9 0 7.65685 0 6V3C0 1.34315 1.34315 0 3 0ZM14 16H21C22.6569 16 24 17.3431 24 19V21C24 22.6569 22.6569 24 21 24H14C12.3431 24 11 22.6569 11 21V19C11 17.3431 12.3431 16 14 16ZM14 18C13.4477 18 13 18.4477 13 19V21C13 21.5523 13.4477 22 14 22H21C21.5523 22 22 21.5523 22 21V19C22 18.4477 21.5523 18 21 18H14ZM3 11H6C7.65685 11 9 12.3431 9 14V21C9 22.6569 7.65685 24 6 24H3C1.34315 24 0 22.6569 0 21V14C0 12.3431 1.34315 11 3 11ZM3 13C2.44772 13 2 13.4477 2 14V21C2 21.5523 2.44772 22 3 22H6C6.55228 22 7 21.5523 7 21V14C7 13.4477 6.55228 13 6 13H3ZM21 0C22.6569 0 24 1.34315 24 3V11C24 12.6569 22.6569 14 21 14H14C12.3431 14 11 12.6569 11 11V3C11 1.34315 12.3431 0 14 0H21ZM13 3V11C13 11.5523 13.4477 12 14 12H21C21.5523 12 22 11.5523 22 11V3C22 2.44772 21.5523 2 21 2H14C13.4477 2 13 2.44772 13 3Z"
      fill="currentColor"
    ></path>
  </svg>
));

// ================= Credits ================== //
export const Credits = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className}
  >
    <circle cx="12" cy="12" r="10" />
    <path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" />
    <path d="M12 18V6" />
  </svg>
));

// ================ Display Name ================ //
User.displayName = "User";
Key.displayName = "Key";
Lock.displayName = "Lock";
Dashboard.displayName = "Dashboard";
Credits.displayName = "Credits";
