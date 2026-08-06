import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import "./index.css";
import App from "./App.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    {/* import.meta.env.BASE_URL reflete `base` do vite.config.ts — '/' em
        dev, '/yamabiko/' no build de produção (ver web/Dockerfile.prod).
        Sem isso as rotas do React Router ignorariam o prefixo e navegação
        interna (Link/navigate) quebraria same-origin em /yamabiko/*. */}
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
