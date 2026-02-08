import { Routes, Route } from "react-router-dom";
import { lazy, Suspense } from "react";
import { Loader } from "@/common/Loader";

// LAYOUTS
const AuthLayout = lazy(() => import("./layouts/AuthLayout"));

// AUTH PAGES
const Login = lazy(() => import("./pages/auth/Login"));

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
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  )
}

export default App