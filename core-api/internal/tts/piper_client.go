package tts

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

// PiperClient fala Wyoming, o protocolo usado pelo Piper (motor de TTS em
// inglês open-source) quando roda como serviço de rede — TCP puro, sem API
// HTTP (confirmado empiricamente contra uma instância real: não responde
// nenhuma requisição HTTP). Cada mensagem tem 3 partes: uma linha JSON de
// cabeçalho (`type`, `data_length`, `payload_length` opcional), seguida de
// exatos `data_length` bytes de um segundo blob JSON (sem newline), seguida
// de exatos `payload_length` bytes de payload binário quando presente — o
// Piper entrega áudio como PCM cru nesse payload, sem cabeçalho WAV, então
// este client remonta o WAV a partir do formato anunciado em "audio-start".
type PiperClient struct {
	address     string // host:port — endereço TCP, não uma URL HTTP
	dialTimeout time.Duration
}

func NewPiperClient(address string) *PiperClient {
	return &PiperClient{address: address, dialTimeout: 10 * time.Second}
}

func (c *PiperClient) Synthesize(ctx context.Context, text, providerVoiceID string) ([]byte, error) {
	dialer := net.Dialer{Timeout: c.dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return nil, fmt.Errorf("piper: falha ao conectar em %s: %w", c.address, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if err := c.sendSynthesize(conn, text, providerVoiceID); err != nil {
		return nil, fmt.Errorf("piper: falha ao enviar synthesize: %w", err)
	}

	audio, err := receiveWyomingAudio(conn)
	if err != nil {
		return nil, fmt.Errorf("piper: %w", err)
	}
	return audio, nil
}

func (c *PiperClient) sendSynthesize(conn net.Conn, text, providerVoiceID string) error {
	data := map[string]any{"text": text}
	if providerVoiceID != "" {
		data["voice"] = map[string]string{"name": providerVoiceID}
	}

	encoded, err := json.Marshal(map[string]any{"type": "synthesize", "data": data})
	if err != nil {
		return err
	}
	_, err = conn.Write(append(encoded, '\n'))
	return err
}

type wyomingHeader struct {
	Type          string `json:"type"`
	DataLength    int    `json:"data_length"`
	PayloadLength int    `json:"payload_length"`
}

type wyomingAudioFormat struct {
	Rate     int `json:"rate"`
	Width    int `json:"width"`
	Channels int `json:"channels"`
}

// receiveWyomingAudio lê a sequência audio-start -> N*audio-chunk ->
// audio-stop e remonta um WAV a partir do PCM bruto recebido.
func receiveWyomingAudio(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)
	var format wyomingAudioFormat
	var pcm bytes.Buffer
	haveFormat := false

	for {
		header, data, payload, err := readWyomingMessage(reader)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler mensagem: %w", err)
		}

		switch header.Type {
		case "audio-start":
			if err := json.Unmarshal(data, &format); err != nil {
				return nil, fmt.Errorf("audio-start inválido: %w", err)
			}
			haveFormat = true
		case "audio-chunk":
			pcm.Write(payload)
		case "audio-stop":
			if !haveFormat {
				return nil, fmt.Errorf("audio-stop recebido sem audio-start prévio")
			}
			return pcmToWav(pcm.Bytes(), format.Rate, format.Width, format.Channels), nil
		case "error":
			return nil, fmt.Errorf("servidor retornou erro: %s", data)
		}
	}
}

func readWyomingMessage(reader *bufio.Reader) (wyomingHeader, []byte, []byte, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return wyomingHeader{}, nil, nil, err
	}

	var header wyomingHeader
	if err := json.Unmarshal(line, &header); err != nil {
		return wyomingHeader{}, nil, nil, fmt.Errorf("cabeçalho inválido (%q): %w", line, err)
	}

	var data []byte
	if header.DataLength > 0 {
		data = make([]byte, header.DataLength)
		if _, err := io.ReadFull(reader, data); err != nil {
			return header, nil, nil, err
		}
	}

	var payload []byte
	if header.PayloadLength > 0 {
		payload = make([]byte, header.PayloadLength)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return header, data, nil, err
		}
	}

	return header, data, payload, nil
}

// pcmToWav monta um cabeçalho WAV (PCM linear) na frente do áudio cru — o
// Wyoming entrega só o PCM, sem envelope de arquivo.
func pcmToWav(pcm []byte, rate, width, channels int) []byte {
	dataSize := len(pcm)
	byteRate := rate * channels * width
	blockAlign := channels * width
	bitsPerSample := width * 8

	buf := bytes.NewBuffer(make([]byte, 0, 44+dataSize))
	buf.WriteString("RIFF")
	writeUint32LE(buf, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	writeUint32LE(buf, 16) // Subchunk1Size (PCM)
	writeUint16LE(buf, 1)  // AudioFormat = PCM
	writeUint16LE(buf, uint16(channels))
	writeUint32LE(buf, uint32(rate))
	writeUint32LE(buf, uint32(byteRate))
	writeUint16LE(buf, uint16(blockAlign))
	writeUint16LE(buf, uint16(bitsPerSample))
	buf.WriteString("data")
	writeUint32LE(buf, uint32(dataSize))
	buf.Write(pcm)
	return buf.Bytes()
}

func writeUint32LE(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func writeUint16LE(buf *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}
