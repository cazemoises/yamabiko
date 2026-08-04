-- Importação de seed/english_curriculum_seed.json (30 exercícios, categorias
-- saudacoes/compras/restaurante/direcoes/emergencia_saude/trabalho_social),
-- language='en-US'. Nenhuma colisão de dedupe possível contra o currículo
-- ja-JP existente: expected_transcript não é UNIQUE no schema (dedupe das
-- migrations anteriores foi feito manualmente por análise, não por
-- constraint), e os dois idiomas nunca compartilham o mesmo transcript.
-- sprint_day_ref 1-30 sequencial: diferente do currículo ja-JP (que ocupa o
-- "dia 1-60" do sprint principal do produto, Sec. 1 do CLAUDE.md), o inglês é
-- uma trilha secundária — o frontend agrupa por dia *depois* de filtrar por
-- idioma (toggle JA/EN), então a numeração não colide visualmente com a
-- trilha japonesa.
-- expected_romaji fica NULL (não aplicável a inglês, coluna já é nullable).
INSERT INTO exercises (category, difficulty, prompt_pt, expected_transcript, sprint_day_ref, language) VALUES
('saudacoes', 1, 'Cumprimente alguém casualmente dizendo ''oi, como você está?''', 'hi, how are you?', 1, 'en-US'),
('saudacoes', 1, 'Responda que está bem, obrigado, e pergunte de volta.', 'I''m good, thanks. And you?', 2, 'en-US'),
('saudacoes', 1, 'Apresente-se dizendo seu nome.', 'my name is Moisés', 3, 'en-US'),
('saudacoes', 2, 'Diga que é um prazer te conhecer.', 'nice to meet you', 4, 'en-US'),
('saudacoes', 2, 'Se despeça dizendo que precisa ir e foi bom te ver.', 'I have to go, it was good seeing you', 5, 'en-US'),
('compras', 1, 'Pergunte quanto custa um item.', 'how much does this cost?', 6, 'en-US'),
('compras', 2, 'Diga que só está olhando, quando o vendedor perguntar se precisa de ajuda.', 'I''m just looking, thanks', 7, 'en-US'),
('compras', 2, 'Pergunte se eles aceitam cartão de crédito.', 'do you accept credit cards?', 8, 'en-US'),
('compras', 3, 'Diga que quer trocar um item porque veio com defeito.', 'I''d like to exchange this, it''s defective', 9, 'en-US'),
('compras', 2, 'Pergunte onde fica o provador.', 'where is the fitting room?', 10, 'en-US'),
('restaurante', 1, 'Peça uma mesa para duas pessoas.', 'a table for two, please', 11, 'en-US'),
('restaurante', 2, 'Pergunte o que o garçom recomenda.', 'what do you recommend?', 12, 'en-US'),
('restaurante', 2, 'Peça a conta.', 'can I have the check, please?', 13, 'en-US'),
('restaurante', 3, 'Diga que é alérgico a amendoim.', 'I''m allergic to peanuts', 14, 'en-US'),
('restaurante', 2, 'Peça um copo d''água sem gás.', 'a glass of still water, please', 15, 'en-US'),
('direcoes', 2, 'Pergunte como chegar na estação de metrô mais próxima.', 'how do I get to the nearest subway station?', 16, 'en-US'),
('direcoes', 1, 'Pergunte se fica perto daqui.', 'is it close to here?', 17, 'en-US'),
('direcoes', 3, 'Diga que está perdido e precisa de ajuda.', 'I''m lost, can you help me?', 18, 'en-US'),
('direcoes', 2, 'Pergunte se pode chamar um táxi.', 'can you call me a taxi?', 19, 'en-US'),
('direcoes', 2, 'Pergunte quanto tempo demora a pé.', 'how long does it take on foot?', 20, 'en-US'),
('emergencia_saude', 3, 'Diga que precisa de um médico.', 'I need a doctor', 21, 'en-US'),
('emergencia_saude', 2, 'Diga que não está se sentindo bem.', 'I''m not feeling well', 22, 'en-US'),
('emergencia_saude', 3, 'Explique que perdeu seu passaporte.', 'I lost my passport', 23, 'en-US'),
('emergencia_saude', 3, 'Peça pra chamar a polícia.', 'please call the police', 24, 'en-US'),
('emergencia_saude', 2, 'Pergunte onde fica o hospital mais próximo.', 'where is the nearest hospital?', 25, 'en-US'),
('trabalho_social', 3, 'Se apresente profissionalmente dizendo o que você faz.', 'I work as a software developer', 26, 'en-US'),
('trabalho_social', 2, 'Diga que sente muito por chegar atrasado.', 'sorry I''m late', 27, 'en-US'),
('trabalho_social', 3, 'Pergunte se pode remarcar a reunião.', 'can we reschedule the meeting?', 28, 'en-US'),
('trabalho_social', 2, 'Diga que não entendeu e peça pra repetir.', 'sorry, could you repeat that?', 29, 'en-US'),
('trabalho_social', 4, 'Diga que vai pensar sobre a proposta e retornar em breve.', 'I''ll think about the offer and get back to you soon', 30, 'en-US');
