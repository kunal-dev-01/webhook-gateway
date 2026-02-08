import { memo } from "react";

// ================ Check ================ //
export const Check = memo(
  ({ className = "w-5 h-5 text-emerald-400" }: { className?: string }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      width="100%"
      height="100%"
      fill="currentColor"
      className={className}
    >
      <path
        fillRule="evenodd"
        d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12zm13.36-1.814a.75.75 0 10-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z"
        clipRule="evenodd"
      />
    </svg>
  ),
);

// ================ X ================ //
export const X = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="100%"
    height="100%"
    viewBox="0 0 24 24"
    className={className}
  >
    <path
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.5"
      d="m5 19l7-7m0 0l7-7m-7 7L5 5m7 7l7 7"
    />
  </svg>
));

// ================ Download ================ //
export const Download = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="100%"
    height="100%"
    viewBox="0 0 24 24"
    className={className}
  >
    <path
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2M7 11l5 5l5-5m-5-7v12"
    />
  </svg>
));

// ================ Error ================ //
export const Error = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="100%"
    height="100%"
    viewBox="0 0 16 16"
    className={className}
  >
    <path
      fill="currentColor"
      fillRule="evenodd"
      d="M8 14.5a6.5 6.5 0 1 0 0-13a6.5 6.5 0 0 0 0 13ZM8 16A8 8 0 1 0 8 0a8 8 0 0 0 0 16Zm1-5a1 1 0 1 1-2 0a1 1 0 0 1 2 0Zm-.25-6.25a.75.75 0 0 0-1.5 0v3.5a.75.75 0 0 0 1.5 0v-3.5Z"
      clipRule="evenodd"
    />
  </svg>
));

// ================= Trash ================= //
export const Trash = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="100%"
    height="100%"
    viewBox="0 0 24 24"
    className={className}
  >
    <path
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      d="M4 7h16m-10 4v6m4-6v6M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3"
    />
  </svg>
));

// ================= ChevronDown ================= //
export const ChevronDown = memo(
  ({ className = "w-5 h-5" }: { className?: string }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="100%"
      height="100%"
      viewBox="0 0 16 16"
      fill="none"
      className={className}
    >
      <path
        d="M4 6L8 10L12 6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  ),
);

// ================= Settings ================= //
export const Settings = memo(
  ({
    className = "w-5 h-5 text-zinc-600 dark:text-zinc-400",
  }: {
    className?: string;
  }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="100%"
      height="100%"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
    >
      <g
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
      >
        <path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 0 0-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 0 0-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 0 0-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 0 0-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 0 0 1.066-2.573c-.94-1.543.826-3.31 2.37-2.37c1 .608 2.296.07 2.572-1.065z" />
        <path d="M9 12a3 3 0 1 0 6 0a3 3 0 0 0-6 0" />
      </g>
    </svg>
  ),
);

// ================= CopyIcon ================= //
export const Copy = memo(
  ({
    className = "w-5 h-5 text-zinc-600 dark:text-zinc-400",
  }: {
    className?: string;
  }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="100%"
      height="100%"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
    >
      <g
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.5"
      >
        <path d="M16.75 5.75a3 3 0 0 0-3-3h-6.5a3 3 0 0 0-3 3v9.5a3 3 0 0 0 3 3h6.5a3 3 0 0 0 3-3z" />
        <path d="M19.75 6.75v8.5a6 6 0 0 1-6 6h-5.5" />
      </g>
    </svg>
  ),
);

// ================= TriangleAlert ================= //
export const TriangleAlert = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    stroke-width="1.5"
    stroke="currentColor"
    className={className}
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
    />
  </svg>
));

// ================= QuestionMark ================= //
export const QuestionMark = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    stroke-width="1.5"
    stroke="currentColor"
    className={className}
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 5.25h.008v.008H12v-.008Z"
    />
  </svg>
));

// ================= Sidebar ================= //
export const Sidebar = memo(({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="100%"
    height="100%"
    viewBox="0 0 20 20"
    className={className}
  >
    <path
      fill="currentColor"
      fillRule="evenodd"
      d="M3 5a1 1 0 0 1 1-1h12a1 1 0 1 1 0 2H4a1 1 0 0 1-1-1Zm0 5a1 1 0 0 1 1-1h6a1 1 0 1 1 0 2H4a1 1 0 0 1-1-1Zm0 5a1 1 0 0 1 1-1h12a1 1 0 1 1 0 2H4a1 1 0 0 1-1-1Z"
      clipRule="evenodd"
    />
  </svg>
));

// ================= SocialLinks ================= //
export const SOCIAL_LINKS = [
  {
    name: "Twitter",
    icon: "M23 3a10.9 10.9 0 01-3.14 1.53 4.48 4.48 0 00-7.86 3v1A10.66 10.66 0 013 4s-4 9 5 13a11.64 11.64 0 01-7 2c9 5 20 0 20-11.5a4.5 4.5 0 00-.08-.83A7.72 7.72 0 0023 3z",
    href: "https://twitter.com/segment",
  },
  {
    name: "GitHub",
    icon: "M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 00-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0020 4.77 5.07 5.07 0 0019.91 1S18.73.65 16 2.48a13.38 13.38 0 00-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 005 4.77a5.44 5.44 0 00-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 009 18.13V22",
    href: "https://github.com/segment",
  },
  {
    name: "Discord",
    icon: "M20.317 4.37a19.791 19.791 0 00-4.885-1.515.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.736 19.736 0 003.677 4.37a.07.07 0 00-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 00.031.057 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 00-.041-.106 13.107 13.107 0 01-1.872-.892.077.077 0 01-.008-.128 10.2 10.2 0 00.372-.292.074.074 0 01.077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.01c.12.098.246.198.373.292a.077.077 0 01-.006.127 12.299 12.299 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 00.084.028 19.839 19.839 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 00-.031-.03z",
    href: "https://discord.com/invite/segment",
  },
];

// ================= EyeOpen ================= //
export const EyeOpen = memo(
  ({ className = "w-5 h-5" }: { className?: string }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="100%"
      height="100%"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
    >
      <g fill="currentColor" fillRule="evenodd" clipRule="evenodd">
        <path d="M16 12a4 4 0 1 1-8 0a4 4 0 0 1 8 0Zm-2 0a2 2 0 1 1-4 0a2 2 0 0 1 4 0Z" />
        <path d="M12 3c5.591 0 10.29 3.824 11.622 9c-1.332 5.176-6.03 9-11.622 9S1.71 17.176.378 12C1.71 6.824 6.408 3 12 3Zm0 16c-4.476 0-8.269-2.942-9.543-7C3.731 7.942 7.524 5 12 5c4.476 0 8.269 2.942 9.543 7c-1.274 4.058-5.067 7-9.543 7Z" />
      </g>
    </svg>
  )
);

// ================= EyeClose ================= //
export const EyeClose = memo(
  ({ className = "w-5 h-5" }: { className?: string }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="200"
      height="200"
      viewBox="0 0 32 32"
      fill="currentColor"
      className={className}
    >
      <path
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        d="M21 16a5 5 0 0 1-5 5m-5-5a5 5 0 0 1 5-5m-3 13.654A13.38 13.38 0 0 0 16 25c7.18 0 13-6 13-9c0-1.336-1.155-3.268-3.071-5M19.5 7.47A13.49 13.49 0 0 0 16 7C8.82 7 3 13 3 16c0 1.32 1.127 3.22 3 4.935M7 25L25 7"
      />
    </svg>
  )
);

// ================= CircleCheck ================= //
export const CircleCheck = memo(
  ({ className = "w-5 h-5" }: { className?: string }) => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="100%"
      height="100%"
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  )
);

// ================= Mail ================= //
export const Mail = memo(
  ({ className = "w-10 h-10 text-indigo-500" }: { className?: string }) => (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="2"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
      />
    </svg>
  )
);

// ========================= Refresh =========================
export const Refresh = memo(
  ({ className = "w-5 h-5" }: { className?: string }) => (
    <svg
      width="100%"
      height="100%"
      viewBox="0 0 25 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <path
        d="M3.13644 9.54175C3.02923 9.94185 3.26667 10.3531 3.66676 10.4603C4.06687 10.5675 4.47812 10.3301 4.58533 9.92998C5.04109 8.22904 6.04538 6.72602 7.44243 5.65403C8.83948 4.58203 10.5512 4.00098 12.3122 4.00098C14.0731 4.00098 15.7848 4.58203 17.1819 5.65403C18.3999 6.58866 19.3194 7.85095 19.8371 9.28639L18.162 8.34314C17.801 8.1399 17.3437 8.26774 17.1405 8.62867C16.9372 8.98959 17.0651 9.44694 17.426 9.65017L20.5067 11.3849C20.68 11.4825 20.885 11.5072 21.0766 11.4537C21.2682 11.4001 21.4306 11.2727 21.5282 11.0993L23.2629 8.01828C23.4661 7.65734 23.3382 7.2 22.9773 6.99679C22.6163 6.79358 22.159 6.92145 21.9558 7.28239L21.195 8.63372C20.5715 6.98861 19.5007 5.54258 18.095 4.464C16.436 3.19099 14.4033 2.50098 12.3122 2.50098C10.221 2.50098 8.1883 3.19099 6.52928 4.464C4.87027 5.737 3.67766 7.52186 3.13644 9.54175Z"
        fill="currentColor"
      />
      <path
        d="M21.4906 14.4582C21.5978 14.0581 21.3604 13.6469 20.9603 13.5397C20.5602 13.4325 20.1489 13.6699 20.0417 14.07C19.5859 15.7709 18.5816 17.274 17.1846 18.346C15.7875 19.418 14.0758 19.999 12.3149 19.999C10.5539 19.999 8.84219 19.418 7.44514 18.346C6.2292 17.4129 5.31079 16.1534 4.79261 14.721L6.45529 15.6573C6.81622 15.8605 7.27356 15.7327 7.47679 15.3718C7.68003 15.0108 7.55219 14.5535 7.19127 14.3502L4.11056 12.6155C3.93723 12.5179 3.73222 12.4932 3.54065 12.5467C3.34907 12.6003 3.18662 12.7278 3.08903 12.9011L1.3544 15.9821C1.15119 16.3431 1.27906 16.8004 1.64 17.0036C2.00094 17.2068 2.45828 17.079 2.66149 16.718L3.42822 15.3562C4.05115 17.0054 5.12348 18.4552 6.532 19.536C8.19102 20.809 10.2237 21.499 12.3149 21.499C14.406 21.499 16.4387 20.809 18.0977 19.536C19.7568 18.263 20.9494 16.4781 21.4906 14.4582Z"
        fill="currentColor"
      />
    </svg>
  )
);



Check.displayName = "Check";
X.displayName = "X";
Download.displayName = "Download";
Error.displayName = "Error";
Trash.displayName = "Trash";
ChevronDown.displayName = "ChevronDown";
Settings.displayName = "Settings";
Copy.displayName = "Copy";
TriangleAlert.displayName = "TriangleAlert";
QuestionMark.displayName = "QuestionMark";
Sidebar.displayName = "Sidebar";
EyeOpen.displayName = "EyeOpen";
EyeClose.displayName = "EyeClose";
CircleCheck.displayName = "CircleCheck";
Mail.displayName = "Mail";
Refresh.displayName = "Refresh";


