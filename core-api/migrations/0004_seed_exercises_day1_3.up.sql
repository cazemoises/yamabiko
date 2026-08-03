-- Seed: exercícios dos dias 1-3 do sprint de 60 dias.
-- Formato pedagógico interno (Sec. 10 CLAUDE.md): [Kana] (Pronúncia PT-BR) {Lógica/Significado}.
-- Zero kanji, zero gramática literária — só o essencial pra sobrevivência conversacional.

INSERT INTO exercises (category, difficulty, prompt_pt, expected_transcript, expected_romaji, sprint_day_ref) VALUES
-- Dia 1 — saudação
-- [おはよう] (Ohaiô) {Bom dia — informal, usado até por volta do meio-dia}
('saudacao', 1, 'Cumprimente alguém dizendo "bom dia" de forma casual.', 'おはよう', 'ohayou', 1),
-- [こんにちは] (Konnitiwa) {Olá / boa tarde — saudação neutra genérica}
('saudacao', 1, 'Diga "olá" / "boa tarde" de forma neutra.', 'こんにちは', 'konnichiwa', 1),
-- [ありがとうございます] (Arigatô gozaimás) {Muito obrigado — formal e educado}
('saudacao', 2, 'Agradeça de forma educada, dizendo "muito obrigado".', 'ありがとうございます', 'arigatou gozaimasu', 1),

-- Dia 2 — auto-apresentação
-- [わたしのなまえはモイゼスです] (Uatashi no namáe wa Moizesu dêss) {Fórmula de auto-apresentação: "meu nome é Moisés"}
('apresentacao', 2, 'Diga seu nome se apresentando: "meu nome é Moisés".', 'わたしのなまえはモイゼスです', 'watashi no namae wa moisesu desu', 2),
-- [ブラジルからきました] (Burajíru kara kimashita) {"Vim do Brasil" — de onde você é}
('apresentacao', 2, 'Diga que você veio do Brasil.', 'ブラジルからきました', 'burajiru kara kimashita', 2),
-- [よろしくおねがいします] (Iorosshiku onegaishimás) {"Prazer em conhecê-lo" — fecha uma apresentação}
('apresentacao', 1, 'Diga "prazer em conhecê-lo" ao final de uma apresentação.', 'よろしくおねがいします', 'yoroshiku onegaishimasu', 2),

-- Dia 3 — konbini básico
-- [これはいくらですか] (Koré wa íkura dêsska) {"Quanto custa isso?" — pergunta de preço}
('konbini', 2, 'Pergunte "quanto custa isso?" no caixa.', 'これはいくらですか', 'kore wa ikura desu ka', 3),
-- [ふくろをください] (Fukuro o kudassai) {"Uma sacola, por favor" — pedido simples}
('konbini', 1, 'Peça uma sacola plástica educadamente.', 'ふくろをください', 'fukuro wo kudasai', 3),
-- [カードではらえますか] (Kaado de haraemásska) {"Posso pagar com cartão?"}
('konbini', 2, 'Pergunte se pode pagar com cartão.', 'カードではらえますか', 'kaado de haraemasu ka', 3),
-- [だいじょうぶです] (Daijôbu dêss) {"Não precisa, obrigado" — recusa educada, ex: sacola extra}
('konbini', 1, 'Recuse algo educadamente, dizendo "não precisa, obrigado".', 'だいじょうぶです', 'daijoubu desu', 3);
