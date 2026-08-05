import { api } from "../../lib/apiClient";

export interface Profile {
  id: string;
  email: string;
  name: string;
  created_at: string;
  current_sprint_day: number;
  xp_total: number;
  current_streak_days: number;
  longest_streak_days: number;
  last_attempt_date?: string;
  badges: string[];
  preferred_voice_ja?: string;
  preferred_voice_en?: string;
}

export function getProfile(): Promise<Profile> {
  return api.get<Profile>("/users/me");
}

// voiceId="" limpa a preferência salva (volta pro default do idioma).
export function updateVoicePreference(language: string, voiceId: string): Promise<void> {
  return api.patch<void>("/users/me/voice-preference", { language, voice_id: voiceId });
}
