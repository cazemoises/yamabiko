import type { ReactNode } from "react";
import { Link, Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider, useAuth } from "./features/auth/AuthContext";
import { LoginPage } from "./features/auth/LoginPage";
import { RegisterPage } from "./features/auth/RegisterPage";
import { ExercisesListPage } from "./features/exercises/ExercisesListPage";
import { ExercisePage } from "./features/exercises/ExercisePage";
import { DashboardPage } from "./features/dashboard/DashboardPage";

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function NavBar() {
  const { isAuthenticated, logout } = useAuth();
  if (!isAuthenticated) return null;
  return (
    <nav className="navbar">
      <span className="brand">やまびこ</span>
      <Link to="/exercises">Exercícios</Link>
      <Link to="/dashboard">Progresso</Link>
      <button type="button" onClick={logout}>
        Sair
      </button>
    </nav>
  );
}

function AppRoutes() {
  return (
    <>
      <NavBar />
      <main>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/exercises"
            element={
              <ProtectedRoute>
                <ExercisesListPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/exercises/:id"
            element={
              <ProtectedRoute>
                <ExercisePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<Navigate to="/exercises" replace />} />
        </Routes>
      </main>
    </>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}
