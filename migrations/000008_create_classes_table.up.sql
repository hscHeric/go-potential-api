-- Tipo de status da aula
CREATE TYPE class_status AS ENUM ('scheduled', 'completed', 'cancelled', 'no_show');

-- Tabela de aulas agendadas
CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    time_slot_id UUID REFERENCES time_slots(id) ON DELETE SET NULL,
    scheduled_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status class_status DEFAULT 'scheduled',
    title VARCHAR(255),
    description TEXT,
    class_link TEXT,
    material_id UUID REFERENCES files(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES auth(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_class_time CHECK (end_time > start_time)
);

-- Índices
CREATE INDEX idx_classes_teacher_id ON classes(teacher_id);
CREATE INDEX idx_classes_scheduled_date ON classes(scheduled_date);
CREATE INDEX idx_classes_status ON classes(status);
CREATE INDEX idx_classes_time_slot_id ON classes(time_slot_id);

-- Trigger para updated_at
CREATE TRIGGER update_classes_updated_at
    BEFORE UPDATE ON classes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comentários
COMMENT ON TABLE classes IS 'Aulas agendadas (podem ter múltiplos alunos)';
COMMENT ON COLUMN classes.title IS 'Título/tema da aula';
COMMENT ON COLUMN classes.description IS 'Descrição/conteúdo da aula';
COMMENT ON COLUMN classes.class_link IS 'Link para sala de aula online (Zoom, Meet, etc)';
COMMENT ON COLUMN classes.created_by IS 'Quem criou a aula (professor ou admin)';
