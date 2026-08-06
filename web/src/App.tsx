import type { ReactNode } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/layout/AppShell";
import { AuthProvider, useAuth } from "./features/auth/AuthContext";
import { AppearanceProvider } from "./features/users/AppearanceContext";
import { LoginPage } from "./features/auth/LoginPage";
import { ProfileSelectPage } from "./features/auth/ProfileSelectPage";
import { RegisterPage } from "./features/auth/RegisterPage";
import { HomePage } from "./features/home/HomePage";
import { ExercisesListPage } from "./features/exercises/ExercisesListPage";
import { ExercisePage } from "./features/exercises/ExercisePage";
import { ScenariosListPage } from "./features/scenarios/ScenariosListPage";
import { ScenarioPage } from "./features/scenarios/ScenarioPage";
import { DashboardPage } from "./features/dashboard/DashboardPage";
import { VoiceSettingsPage } from "./features/settings/VoiceSettingsPage";

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function AppRoutes() {
  return (
    <AppShell>
      <Routes>
        <Route path="/login" element={<ProfileSelectPage />} />
        <Route path="/login/password" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <HomePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/scenarios"
          element={
            <ProtectedRoute>
              <ScenariosListPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/scenarios/:id"
          element={
            <ProtectedRoute>
              <ScenarioPage />
            </ProtectedRoute>
          }
        />
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
        <Route
          path="/settings/voice"
          element={
            <ProtectedRoute>
              <VoiceSettingsPage />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppearanceProvider>
        <AppRoutes />
      </AppearanceProvider>
    </AuthProvider>
  );
}
