-- Tabela de horários disponíveis do professor (recorrentes)
CREATE TABLE time_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    max_students INTEGER NOT NULL DEFAULT 1,
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT valid_max_students CHECK (max_students > 0)
);

-- Índices
CREATE INDEX idx_time_slots_teacher_id ON time_slots(teacher_id);
CREATE INDEX idx_time_slots_day_of_week ON time_slots(day_of_week);
CREATE INDEX idx_time_slots_is_available ON time_slots(is_available);

-- Trigger para updated_at
CREATE TRIGGER update_time_slots_updated_at
    BEFORE UPDATE ON time_slots
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comentários
COMMENT ON TABLE time_slots IS 'Horários disponíveis recorrentes do professor';
COMMENT ON COLUMN time_slots.day_of_week IS 'Dia da semana (0=Domingo, 1=Segunda, ..., 6=Sábado)';
COMMENT ON COLUMN time_slots.max_students IS 'Número máximo de alunos por horário (1=individual, >1=grupo)';
