import { Navigate, useLocation } from "react-router-dom";
import { useAuthStore } from "@/lib/auth";

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRoles?: string[];
}

export default function ProtectedRoute({
  children,
  requiredRoles,
}: ProtectedRouteProps) {
  const { isAuthenticated, user } = useAuthStore();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requiredRoles && user && !requiredRoles.some((r) => user.roles?.includes(r))) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}
