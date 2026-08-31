import type { ReactNode } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/layout/AppShell";
import { AuthProvider, useAuth } from "./features/auth/AuthContext";
import { AppearanceProvider } from "./features/users/AppearanceContext";
import { HomePage } from "./features/home/HomePage";
import { ExercisesListPage } from "./features/exercises/ExercisesListPage";
import { ExercisePage } from "./features/exercises/ExercisePage";
import { ScenariosListPage } from "./features/scenarios/ScenariosListPage";
import { ScenarioPage } from "./features/scenarios/ScenarioPage";
import { DashboardPage } from "./features/dashboard/DashboardPage";
import { VoiceSettingsPage } from "./features/settings/VoiceSettingsPage";

// Sem login local: "checking" mostra um loading breve enquanto GET /users/me
// resolve, "unauthenticated" significa que a requisição não chegou com um
// Remote-Email válido (acesso direto contornando o Pangolin, ou Pangolin
// mal configurado) — não tem pra onde redirecionar, só explicar (ver
// features/auth/AuthContext.tsx).
function ProtectedRoute({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  if (status === "checking") {
    return <p className="center-message">Carregando...</p>;
  }
  if (status === "unauthenticated") {
    return (
      <p className="center-message">
        Não foi possível confirmar sua identidade. Acesse o やまびこ pelo Pangolin.
      </p>
    );
  }
  return <>{children}</>;
}

function AppRoutes() {
  return (
    <AppShell>
      <Routes>
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
