-- Tabela de relacionamento entre aulas e alunos (N:N)
CREATE TABLE class_students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by UUID NOT NULL REFERENCES auth(id),
    attended BOOLEAN DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_class_student UNIQUE (class_id, student_id)
);

-- Índices
CREATE INDEX idx_class_students_class_id ON class_students(class_id);
CREATE INDEX idx_class_students_student_id ON class_students(student_id);

-- Comentários
COMMENT ON TABLE class_students IS 'Alunos matriculados em cada aula';
COMMENT ON COLUMN class_students.added_by IS 'Quem adicionou o aluno (professor, admin ou próprio aluno)';
COMMENT ON COLUMN class_students.attended IS 'Se o aluno compareceu (NULL=não marcado, true=presente, false=ausente)';
