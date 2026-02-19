import { Routes, Route } from "react-router-dom";
import { lazy, Suspense } from "react";
import { Loader } from "@/common/Loader";

// LAYOUTS
const AuthLayout = lazy(() => import("./layouts/AuthLayout"));
const DashboardLayout = lazy(() => import("./layouts/DashboardLayout"));

// AUTH PAGES
const Login = lazy(() => import("./pages/auth/Login"));
const Register = lazy(() => import("./pages/auth/Register"));
const VerifyEmail = lazy(() => import("./pages/auth/VerifyEmail"));
const ForgotPassword = lazy(() => import("./pages/auth/ForgotPassword"));
const ResetPassword = lazy(() => import("./pages/auth/ResetPassword"));

// DASHBOARD PAGES
const Dashboard = lazy(() => import("./pages/dashboard/Dashboard"));

// ORGANIZATION PAGES
const Organization = lazy(() => import("./pages/organization/Organization"));

// WEBHOOKS PAGES
const Webhooks = lazy(() => import("./pages/webhooks/Webhooks"));

// SETTINGS PAGES
const Settings = lazy(() => import("./pages/settings/Settings"));

// NOT FOUND
const NotFound = lazy(() => import("./pages/notfound/NotFound"));

const App = () => {
  return (
    <Suspense
      fallback={
        <div className="w-full h-screen flex items-center justify-center">
          <Loader className="w-8 h-8 text-indigo-500" />
        </div>
      }
    >
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path="/auth/login" element={<Login />} />
          <Route path="/auth/register" element={<Register />} />
          <Route path="/auth/verify-email" element={<VerifyEmail />} />
          <Route path="/auth/forgot-password" element={<ForgotPassword />} />
          <Route path="/auth/reset-password" element={<ResetPassword />} />
        </Route>

        <Route element={<DashboardLayout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/organization" element={<Organization />} />
          <Route path="/webhooks" element={<Webhooks />} />
          <Route path="/settings" element={<Settings />} />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  )
}

export default App