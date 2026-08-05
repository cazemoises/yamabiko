import { api } from "../../lib/apiClient";

export interface Voice {
  id: string;
  name: string;
  language: string;
}

export function listVoices(language: string): Promise<Voice[]> {
  return api.get<Voice[]>(`/tts/voices?language=${encodeURIComponent(language)}`);
}

export function getVoicePreview(voiceId: string): Promise<Blob> {
  return api.getBlob(`/tts/voices/${voiceId}/preview`);
}
