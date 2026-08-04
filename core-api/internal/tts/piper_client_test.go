package tts_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/yamabiko/core-api/internal/tts"
)

// startFakeWyomingServer sobe um servidor TCP mínimo que fala o protocolo
// Wyoming (framing verificado empiricamente contra uma instância real de
// wyoming-piper antes deste commit) — aceita 1 conexão, entrega a mensagem
// "synthesize" recebida pro handler, e escreve de volta o que o handler
// devolver.
func startFakeWyomingServer(t *testing.T, handler func(t *testing.T, synthesizeMsg map[string]any) []byte) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("falha ao abrir listener: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			return
		}

		response := handler(t, msg)
		_, _ = conn.Write(response)
	}()

	return listener.Addr().String()
}

// wyomingMessage monta uma mensagem no formato de 3 partes do Wyoming:
// header JSON (data_length/payload_length) + blob "data" de tamanho exato +
// payload binário de tamanho exato.
func wyomingMessage(msgType string, data map[string]any, payload []byte) []byte {
	dataJSON, _ := json.Marshal(data)
	header := map[string]any{"type": msgType, "data_length": len(dataJSON)}
	if len(payload) > 0 {
		header["payload_length"] = len(payload)
	}
	headerJSON, _ := json.Marshal(header)

	var buf bytes.Buffer
	buf.Write(headerJSON)
	buf.WriteByte('\n')
	buf.Write(dataJSON)
	buf.Write(payload)
	return buf.Bytes()
}

func pcm16Silence(sampleCount int) []byte {
	buf := make([]byte, sampleCount*2)
	for i := range sampleCount {
		binary.LittleEndian.PutUint16(buf[i*2:], 0)
	}
	return buf
}

func TestPiperClient_Synthesize_ReconstructsValidWavFromChunks(t *testing.T) {
	chunk1 := pcm16Silence(100)
	chunk2 := pcm16Silence(50)

	var capturedText string
	addr := startFakeWyomingServer(t, func(t *testing.T, msg map[string]any) []byte {
		data, _ := msg["data"].(map[string]any)
		capturedText, _ = data["text"].(string)

		var out bytes.Buffer
		out.Write(wyomingMessage("audio-start", map[string]any{"rate": 22050, "width": 2, "channels": 1}, nil))
		out.Write(wyomingMessage("audio-chunk", map[string]any{"rate": 22050, "width": 2, "channels": 1}, chunk1))
		out.Write(wyomingMessage("audio-chunk", map[string]any{"rate": 22050, "width": 2, "channels": 1}, chunk2))
		out.Write(wyomingMessage("audio-stop", map[string]any{}, nil))
		return out.Bytes()
	})

	client := tts.NewPiperClient(addr, "en_US-lessac-medium")
	wav, err := client.Synthesize(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedText != "hello world" {
		t.Fatalf("esperava que o servidor recebesse o texto 'hello world', veio %q", capturedText)
	}

	if len(wav) < 44 {
		t.Fatalf("WAV muito curto: %d bytes", len(wav))
	}
	if string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		t.Fatalf("esperava cabeçalho RIFF/WAVE, veio %q/%q", wav[0:4], wav[8:12])
	}
	dataSize := binary.LittleEndian.Uint32(wav[40:44])
	wantDataSize := uint32(len(chunk1) + len(chunk2))
	if dataSize != wantDataSize {
		t.Fatalf("esperava data chunk de %d bytes, veio %d", wantDataSize, dataSize)
	}
	if !bytes.Equal(wav[44:], append(append([]byte{}, chunk1...), chunk2...)) {
		t.Fatal("PCM do WAV reconstruído não bate com os chunks originais")
	}
}

func TestPiperClient_Synthesize_SendsConfiguredVoice(t *testing.T) {
	var capturedVoice string
	addr := startFakeWyomingServer(t, func(t *testing.T, msg map[string]any) []byte {
		data, _ := msg["data"].(map[string]any)
		voice, _ := data["voice"].(map[string]any)
		capturedVoice, _ = voice["name"].(string)

		var out bytes.Buffer
		out.Write(wyomingMessage("audio-start", map[string]any{"rate": 22050, "width": 2, "channels": 1}, nil))
		out.Write(wyomingMessage("audio-stop", map[string]any{}, nil))
		return out.Bytes()
	})

	client := tts.NewPiperClient(addr, "en_US-lessac-medium")
	if _, err := client.Synthesize(context.Background(), "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedVoice != "en_US-lessac-medium" {
		t.Fatalf("esperava voice.name='en_US-lessac-medium' na mensagem enviada, veio %q", capturedVoice)
	}
}

func TestPiperClient_Synthesize_NoVoiceConfigured_OmitsVoiceField(t *testing.T) {
	hadVoiceField := true
	addr := startFakeWyomingServer(t, func(t *testing.T, msg map[string]any) []byte {
		data, _ := msg["data"].(map[string]any)
		_, hadVoiceField = data["voice"]

		var out bytes.Buffer
		out.Write(wyomingMessage("audio-start", map[string]any{"rate": 22050, "width": 2, "channels": 1}, nil))
		out.Write(wyomingMessage("audio-stop", map[string]any{}, nil))
		return out.Bytes()
	})

	client := tts.NewPiperClient(addr, "")
	if _, err := client.Synthesize(context.Background(), "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hadVoiceField {
		t.Fatal("esperava que o campo 'voice' não fosse enviado quando nenhuma voz é configurada")
	}
}

func TestPiperClient_Synthesize_ServerError_ReturnsError(t *testing.T) {
	addr := startFakeWyomingServer(t, func(t *testing.T, msg map[string]any) []byte {
		return wyomingMessage("error", map[string]any{"text": "modelo não encontrado"}, nil)
	})

	client := tts.NewPiperClient(addr, "voz-inexistente")
	if _, err := client.Synthesize(context.Background(), "hi"); err == nil {
		t.Fatal("esperava erro quando o servidor devolve uma mensagem type=error")
	}
}

func TestPiperClient_Synthesize_ConnectionRefused_ReturnsError(t *testing.T) {
	client := tts.NewPiperClient("127.0.0.1:1", "") // porta 1 não deve ter nada escutando
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := client.Synthesize(ctx, "hi"); err == nil {
		t.Fatal("esperava erro ao tentar conectar num endereço sem servidor")
	}
}
