ALTER TABLE phonetic_error_patterns DROP CONSTRAINT phonetic_error_patterns_user_id_pattern_type_language_key;
ALTER TABLE phonetic_error_patterns ADD CONSTRAINT phonetic_error_patterns_user_id_pattern_type_key
    UNIQUE (user_id, pattern_type);

ALTER TABLE phonetic_error_patterns DROP COLUMN language;
ALTER TABLE attempts DROP COLUMN language;
