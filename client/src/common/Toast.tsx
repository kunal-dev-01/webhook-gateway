import { useEffect } from 'react';
import { CircleCheck, Error } from '@/svg/Common';

type ToastProps = {
  status: 'success' | 'error' | null;
  message: string | null | undefined;
  onClose: () => void; // Added onClose prop
};

const Toast = ({ status, message, onClose }: ToastProps) => {
  useEffect(() => {
    if (status && message) {
      // 3 seconds baad auto clear karega
      const timer = setTimeout(() => {
        onClose();
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [status, message, onClose]);

  if (!status || !message) return null;

  const isSuccess = status === 'success';

  return (
    <div className={`flex items-center gap-3 px-4 py-3 rounded-xl mb-4 text-sm font-medium border animate-in slide-in-from-top-2 ${
      isSuccess 
        ? 'bg-emerald-50 border-emerald-100 text-emerald-700 dark:bg-emerald-900/20 dark:border-emerald-900 dark:text-emerald-400' 
        : 'bg-red-50 border-red-100 text-red-700 dark:bg-red-900/20 dark:border-red-900 dark:text-red-400'
    }`}>
      <div className="shrink-0">
        {isSuccess ? <CircleCheck /> : <Error className='w-5 h-5' />}
      </div>
      <p>{message}</p>
    </div>
  );
};

export default Toast;