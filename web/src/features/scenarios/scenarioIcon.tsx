import { Briefcase, HeartPulse, Handshake, MapPin, MessageCircle, Plane, ShoppingBag, UtensilsCrossed, type LucideIcon } from "lucide-react";

// Scenario não expõe category pelo backend (só Exercise tem, ver
// ScenariosListPage) — o título já basta pra escolher um ícone coerente
// pros 9 cenários reais hoje (3 ja-JP + 6 en-US, ver migrations 0014 e
// 0019 do core-api). Ordem importa: regras mais específicas primeiro,
// senão "Cumprimentar um colega no TRABALHO de manhã" bateria com a regra
// de "trabalho" antes da de cumprimentar. Cenário futuro sem match nenhum
// cai no ícone genérico de conversa — nunca quebra, só fica menos específico.
const ICON_RULES: Array<{ pattern: RegExp; icon: LucideIcon }> = [
  { pattern: /cumprimentar|conhecer alguém|se apresentando|apresentar-se/i, icon: Handshake },
  { pattern: /aeroporto|check-in|\bvoo\b|avião/i, icon: Plane },
  { pattern: /jantar|restaurante|comida|refeição/i, icon: UtensilsCrossed },
  { pattern: /compras|\bloja\b/i, icon: ShoppingBag },
  { pattern: /informaç|direç|\brua\b|caminho/i, icon: MapPin },
  { pattern: /emergência|saúde|documento|hospital|médic/i, icon: HeartPulse },
  { pattern: /trabalho|profissiona|reunião|escritório/i, icon: Briefcase },
];

export function iconForScenario(titlePt: string): LucideIcon {
  return ICON_RULES.find(({ pattern }) => pattern.test(titlePt))?.icon ?? MessageCircle;
}
